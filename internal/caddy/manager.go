package caddy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DockerContainer represents a running Docker container
type DockerContainer struct {
	Name  string
	Image string
}

// ListDockerContainers returns a list of running Docker containers
// Optionally filters by image name containing "caddy"
func ListDockerContainers(filterCaddy bool) ([]DockerContainer, error) {
	// Run docker ps to get container names and images
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}|{{.Image}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	var containers []DockerContainer
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		image := strings.TrimSpace(parts[1])

		if filterCaddy {
			// Only include containers with "caddy" in name or image
			lowerName := strings.ToLower(name)
			lowerImage := strings.ToLower(image)
			if !strings.Contains(lowerName, "caddy") && !strings.Contains(lowerImage, "caddy") {
				continue
			}
		}

		containers = append(containers, DockerContainer{
			Name:  name,
			Image: image,
		})
	}

	// Sort by name
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	return containers, nil
}

// BackupCaddyfile creates a timestamped backup of the Caddyfile
func BackupCaddyfile(caddyfilePath string) (string, error) {
	// Get original file permissions
	fileInfo, err := os.Stat(caddyfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat Caddyfile: %w", err)
	}
	originalPerms := fileInfo.Mode().Perm()

	// Read current content
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read Caddyfile: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup.%s", caddyfilePath, timestamp)

	// Write backup with original permissions
	if err := os.WriteFile(backupPath, content, originalPerms); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// AppendEntry adds a new Caddy block to the Caddyfile
func AppendEntry(caddyfilePath string, block string) error {
	// Get original file permissions
	fileInfo, err := os.Stat(caddyfilePath)
	if err != nil {
		return fmt.Errorf("failed to stat Caddyfile: %w", err)
	}
	originalPerms := fileInfo.Mode().Perm()

	// Read current content
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Caddyfile: %w", err)
	}

	// Append new block with spacing
	newContent := string(content)

	// Ensure file ends with newline before adding new block
	if len(newContent) > 0 && newContent[len(newContent)-1] != '\n' {
		newContent += "\n"
	}

	// Add blank line before new block for spacing
	newContent += "\n" + block

	// Write updated content with original permissions
	if err := os.WriteFile(caddyfilePath, []byte(newContent), originalPerms); err != nil {
		return fmt.Errorf("failed to write Caddyfile: %w", err)
	}

	return nil
}

// RemoveEntry removes a Caddy block from the Caddyfile by domain marker or domain line
// Uses brace counting to handle nested blocks correctly
func RemoveEntry(caddyfilePath string, domain string) error {
	// Get original file permissions
	fileInfo, err := os.Stat(caddyfilePath)
	if err != nil {
		return fmt.Errorf("failed to stat Caddyfile: %w", err)
	}
	originalPerms := fileInfo.Mode().Perm()

	// Read current content
	content, err := os.ReadFile(caddyfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Caddyfile: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	newLines := []string{}
	skipBlock := false
	skipMarker := false
	braceCount := 0
	marker := fmt.Sprintf("# === %s ===", domain)

	for i, line := range lines {
		// Check for marker comment (skip this line if found)
		if strings.TrimSpace(line) == marker {
			skipMarker = true
			skipBlock = true
			continue
		}

		// Check for domain line (with or without marker)
		stripped := strings.TrimSpace(line)
		domainPattern1 := domain + " {"
		domainPattern2 := domain + "{"

		if !skipBlock && (stripped == domainPattern1 || stripped == domainPattern2 || strings.HasPrefix(stripped, domain+",") || strings.HasPrefix(stripped, domain+" ")) {
			// Check if this line ends with opening brace
			if strings.HasSuffix(stripped, "{") {
				skipBlock = true
				braceCount = 1
				// Skip marker line if it's on the previous line
				if i > 0 && strings.TrimSpace(lines[i-1]) == marker && !skipMarker {
					// Remove the marker that was already added to newLines
					if len(newLines) > 0 {
						newLines = newLines[:len(newLines)-1]
					}
				}
				continue
			}
		}

		if skipBlock {
			if strings.HasSuffix(stripped, "{") && stripped != "" {
				braceCount++
			}
			if stripped == "}" {
				braceCount--
				if braceCount == 0 {
					skipBlock = false
					skipMarker = false
				}
			}
			continue
		}

		newLines = append(newLines, line)
	}

	// Write updated content with original permissions
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(caddyfilePath, []byte(newContent), originalPerms); err != nil {
		return fmt.Errorf("failed to write Caddyfile: %w", err)
	}

	return nil
}

// RestoreFromBackup restores a Caddyfile from a backup
func RestoreFromBackup(caddyfilePath, backupPath string) error {
	// Get backup file permissions to preserve them
	backupInfo, err := os.Stat(backupPath)
	if err != nil {
		return fmt.Errorf("failed to stat backup: %w", err)
	}
	backupPerms := backupInfo.Mode().Perm()

	// Read backup content
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Restore to original location with backup's permissions
	if err := os.WriteFile(caddyfilePath, content, backupPerms); err != nil {
		return fmt.Errorf("failed to restore Caddyfile: %w", err)
	}

	return nil
}

// FormatCaddyfile runs caddy fmt to format the Caddyfile
// Parameters:
// - caddyfilePath: Host path (for reading/writing)
// - containerPath: Path inside Docker container (for formatting). If empty, uses caddyfilePath
// - containerName: Docker container name
// - dockerMethod: "compose" or "plain" (determines docker compose exec vs docker exec)
// - composeFilePath: Path to docker-compose.yml (for compose method)
func FormatCaddyfile(caddyfilePath, containerPath, containerName, dockerMethod, composeFilePath string) error {
	var cmd *exec.Cmd
	var cmdStr string

	if containerPath != "" {
		// Docker deployment - format inside container
		if dockerMethod == "compose" {
			if composeFilePath != "" {
				cmd = exec.Command("docker", "compose", "-f", composeFilePath, "exec", "-T", containerName, "caddy", "fmt", "--overwrite", containerPath)
				cmdStr = fmt.Sprintf("docker compose -f %s exec -T %s caddy fmt --overwrite %s", composeFilePath, containerName, containerPath)
			} else {
				cmd = exec.Command("docker", "compose", "exec", "-T", containerName, "caddy", "fmt", "--overwrite", containerPath)
				cmdStr = fmt.Sprintf("docker compose exec -T %s caddy fmt --overwrite %s", containerName, containerPath)
			}
		} else {
			cmd = exec.Command("docker", "exec", containerName, "caddy", "fmt", "--overwrite", containerPath)
			cmdStr = fmt.Sprintf("docker exec %s caddy fmt --overwrite %s", containerName, containerPath)
		}
	} else {
		// Local deployment
		cmd = exec.Command("caddy", "fmt", "--overwrite", caddyfilePath)
		cmdStr = fmt.Sprintf("caddy fmt --overwrite %s", caddyfilePath)
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("format command failed: %s\nCommand: %s\nOutput: %s", err, cmdStr, string(output))
	}
	return nil
}

// ValidateCaddyfile runs validation command to check configuration
// Uses custom validation command from config if available, otherwise uses default
// Parameters:
// - caddyfilePath: Host path (for reading/writing)
// - containerPath: Path inside Docker container (for validation). If empty, uses caddyfilePath
// - containerName: Docker container name
// - composeFilePath: Path to docker-compose.yml (for compose method)
// - validationCommand: Custom validation command with {path}, {container}, {compose_file} placeholders
func ValidateCaddyfile(caddyfilePath, containerPath, containerName, composeFilePath, validationCommand string) error {
	var cmd *exec.Cmd
	var cmdStr string

	if validationCommand == "" {
		// Default commands: build exec.Command directly (safe from path splitting issues)
		if containerPath != "" {
			cmd = exec.Command("docker", "compose", "exec", "-T", containerName, "caddy", "validate", "--config", containerPath)
			cmdStr = fmt.Sprintf("docker compose exec -T %s caddy validate --config %s", containerName, containerPath)
		} else {
			cmd = exec.Command("caddy", "validate", "--config", caddyfilePath)
			cmdStr = fmt.Sprintf("caddy validate --config %s", caddyfilePath)
		}
	} else {
		// Custom validation command from user config - use placeholder replacement
		pathForValidation := caddyfilePath
		if containerPath != "" {
			pathForValidation = containerPath
		}
		cmdStr = strings.ReplaceAll(validationCommand, "{path}", pathForValidation)
		cmdStr = strings.ReplaceAll(cmdStr, "{container}", containerName)
		cmdStr = strings.ReplaceAll(cmdStr, "{compose_file}", composeFilePath)

		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			return fmt.Errorf("invalid validation command: empty command")
		}
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("validation command failed: %s\nCommand: %s\nOutput: %s", err, cmdStr, string(output))
	}
	return nil
}

// FormatAndValidateCaddyfile formats and then validates the Caddyfile
// This is the recommended way to check a Caddyfile after modifications
func FormatAndValidateCaddyfile(caddyfilePath, containerPath, containerName, dockerMethod, composeFilePath, validationCommand string) error {
	// Format first (ignore errors - formatting is best-effort)
	_ = FormatCaddyfile(caddyfilePath, containerPath, containerName, dockerMethod, composeFilePath)

	// Validate is required
	return ValidateCaddyfile(caddyfilePath, containerPath, containerName, composeFilePath, validationCommand)
}

// RestartCaddy restarts the Caddy container
func RestartCaddy(containerName string) error {
	cmd := exec.Command("docker", "restart", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restart failed: %s\n%s", err, string(output))
	}
	return nil
}

// BackupInfo holds information about a backup file
type BackupInfo struct {
	Path      string
	Timestamp time.Time
	Size      int64
}

// ListBackups returns a list of all Caddyfile backups, sorted by timestamp (newest first)
func ListBackups(caddyfilePath string) ([]BackupInfo, error) {
	dir := filepath.Dir(caddyfilePath)
	baseName := filepath.Base(caddyfilePath)
	pattern := baseName + ".backup.*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to find backups: %w", err)
	}

	var backups []BackupInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      match,
			Timestamp: info.ModTime(),
			Size:      info.Size(),
		})
	}

	// Sort by timestamp, newest first
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[i].Timestamp.Before(backups[j].Timestamp) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	return backups, nil
}

// GetOldBackups returns a list of backups older than maxAge without deleting them
func GetOldBackups(caddyfilePath string, maxAge time.Duration) ([]BackupInfo, error) {
	dir := filepath.Dir(caddyfilePath)
	baseName := filepath.Base(caddyfilePath)
	pattern := baseName + ".backup.*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("failed to find backups: %w", err)
	}

	var oldBackups []BackupInfo
	now := time.Now()
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			oldBackups = append(oldBackups, BackupInfo{
				Path:      match,
				Timestamp: info.ModTime(),
				Size:      info.Size(),
			})
		}
	}

	// Sort by timestamp, newest first
	for i := 0; i < len(oldBackups)-1; i++ {
		for j := i + 1; j < len(oldBackups); j++ {
			if oldBackups[i].Timestamp.Before(oldBackups[j].Timestamp) {
				oldBackups[i], oldBackups[j] = oldBackups[j], oldBackups[i]
			}
		}
	}

	return oldBackups, nil
}

// CleanupOldBackups removes backup files older than a specified duration
func CleanupOldBackups(caddyfilePath string, maxAge time.Duration) error {
	dir := filepath.Dir(caddyfilePath)
	baseName := filepath.Base(caddyfilePath)
	pattern := baseName + ".backup.*"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return fmt.Errorf("failed to find backups: %w", err)
	}

	now := time.Now()
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			os.Remove(match)
		}
	}

	return nil
}

// CleanupByCount keeps the newest maxCount backups and deletes the rest.
// Returns the number of backups deleted.
func CleanupByCount(caddyfilePath string, maxCount int) (int, error) {
	if maxCount <= 0 {
		return 0, nil
	}

	backups, err := ListBackups(caddyfilePath)
	if err != nil {
		return 0, err
	}

	if len(backups) <= maxCount {
		return 0, nil
	}

	deleted := 0
	// backups are sorted newest first, so delete from maxCount onwards
	for _, b := range backups[maxCount:] {
		if err := os.Remove(b.Path); err == nil {
			deleted++
		}
	}

	return deleted, nil
}

// CleanupBySize removes oldest backups until total size is under maxMB.
// Returns the number of backups deleted.
func CleanupBySize(caddyfilePath string, maxMB int) (int, error) {
	if maxMB <= 0 {
		return 0, nil
	}

	backups, err := ListBackups(caddyfilePath)
	if err != nil {
		return 0, err
	}

	maxBytes := int64(maxMB) * 1024 * 1024
	var totalSize int64
	for _, b := range backups {
		totalSize += b.Size
	}

	if totalSize <= maxBytes {
		return 0, nil
	}

	deleted := 0
	// Delete oldest first (backups sorted newest first, so iterate from end)
	for i := len(backups) - 1; i >= 0 && totalSize > maxBytes; i-- {
		if err := os.Remove(backups[i].Path); err == nil {
			totalSize -= backups[i].Size
			deleted++
		}
	}

	return deleted, nil
}

// GetTotalBackupSize returns the total size of all backups in bytes.
func GetTotalBackupSize(caddyfilePath string) (int64, error) {
	backups, err := ListBackups(caddyfilePath)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, b := range backups {
		total += b.Size
	}
	return total, nil
}
