package config

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetDefaultExportDir returns the default directory for profile exports
func GetDefaultExportDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "lazyproxyflare", "exports"), nil
}

// ExportProfile creates a .tar.gz bundle with the profile YAML and audit log
func ExportProfile(profileName, outputPath string) error {
	// Load the profile to verify it exists
	profileConfig, err := LoadProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Marshal profile to YAML
	profileYAML, err := yaml.Marshal(profileConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Create the tar.gz file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add profile.yaml to archive
	if err := addToTar(tarWriter, "profile.yaml", profileYAML); err != nil {
		return fmt.Errorf("failed to add profile to archive: %w", err)
	}

	// Try to add audit.log if it exists
	homeDir, _ := os.UserHomeDir()
	auditLogPath := filepath.Join(homeDir, ".config", "lazyproxyflare", "audit.log")
	if auditData, err := os.ReadFile(auditLogPath); err == nil {
		if err := addToTar(tarWriter, "audit.log", auditData); err != nil {
			return fmt.Errorf("failed to add audit log to archive: %w", err)
		}
	}

	return nil
}

// ImportProfile extracts a profile from a .tar.gz bundle and saves it
func ImportProfile(archivePath string, overwrite bool) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var profileData []byte
	var profileName string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		data, err := io.ReadAll(tarReader)
		if err != nil {
			return "", fmt.Errorf("failed to read entry data: %w", err)
		}

		switch header.Name {
		case "profile.yaml":
			profileData = data
		}
	}

	if profileData == nil {
		return "", fmt.Errorf("archive does not contain profile.yaml")
	}

	// Parse profile to get the name
	var profileConfig ProfileConfig
	if err := yaml.Unmarshal(profileData, &profileConfig); err != nil {
		return "", fmt.Errorf("failed to parse profile: %w", err)
	}

	profileName = profileConfig.Profile.Name
	if profileName == "" {
		return "", fmt.Errorf("profile has no name")
	}

	// Sanitize profile name
	profileName = sanitizeProfileName(profileName)

	// Check if profile already exists
	if !overwrite {
		existing, _ := ListProfiles()
		for _, name := range existing {
			if strings.EqualFold(name, profileName) {
				return "", fmt.Errorf("profile '%s' already exists (use overwrite to replace)", profileName)
			}
		}
	}

	// Save the profile
	if err := SaveProfile(profileName, &profileConfig); err != nil {
		return "", fmt.Errorf("failed to save imported profile: %w", err)
	}

	return profileName, nil
}

// addToTar adds a file entry to a tar writer
func addToTar(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

// sanitizeProfileName removes characters that aren't safe for filenames
func sanitizeProfileName(name string) string {
	safe := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '_'
	}, name)
	if safe == "" {
		return "imported"
	}
	return safe
}
