package snippet_wizard

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Valid   bool
	Message string
}

// ValidateCIDR validates CIDR notation (e.g., "10.0.0.0/8", "192.168.1.0/24")
func ValidateCIDR(cidr string) ValidationResult {
	if cidr == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return ValidationResult{Valid: false, Message: "Invalid CIDR (e.g., 10.0.0.0/8)"}
	}

	return ValidationResult{Valid: true, Message: "Valid CIDR"}
}

// ValidateIP validates IP address (IPv4 or IPv6)
func ValidateIP(ip string) ValidationResult {
	if ip == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ValidationResult{Valid: false, Message: "Invalid IP address"}
	}

	return ValidationResult{Valid: true, Message: "Valid IP"}
}

// ValidateDuration validates Go duration strings (e.g., "120s", "2m", "1h")
func ValidateDuration(duration string) ValidationResult {
	if duration == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	_, err := time.ParseDuration(duration)
	if err != nil {
		return ValidationResult{Valid: false, Message: "Invalid duration (e.g., 120s, 2m, 1h)"}
	}

	return ValidationResult{Valid: true, Message: "Valid duration"}
}

// ValidateSize validates byte size strings (e.g., "512MB", "1GB", "5GB")
func ValidateSize(size string) ValidationResult {
	if size == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	// Regex: number + unit (KB, MB, GB, TB)
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(KB|MB|GB|TB|K|M|G|T|B)?$`)
	matches := re.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(size)))

	if matches == nil {
		return ValidationResult{Valid: false, Message: "Invalid size (e.g., 512MB, 1GB, 5GB)"}
	}

	// Parse number
	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || num <= 0 {
		return ValidationResult{Valid: false, Message: "Size must be positive"}
	}

	return ValidationResult{Valid: true, Message: "Valid size"}
}

// ValidateIntRange validates integer is within min/max range
func ValidateIntRange(value string, min, max int) ValidationResult {
	if value == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return ValidationResult{Valid: false, Message: "Must be a number"}
	}

	if num < min || num > max {
		return ValidationResult{Valid: false, Message: "Must be between " + strconv.Itoa(min) + " and " + strconv.Itoa(max)}
	}

	return ValidationResult{Valid: true, Message: "Valid"}
}

// ValidatePositiveInt validates integer is positive (> 0)
func ValidatePositiveInt(value string) ValidationResult {
	if value == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return ValidationResult{Valid: false, Message: "Must be a number"}
	}

	if num <= 0 {
		return ValidationResult{Valid: false, Message: "Must be positive"}
	}

	return ValidationResult{Valid: true, Message: "Valid"}
}

// ValidateOrigins validates CORS origins (can be "*" or comma-separated URLs)
func ValidateOrigins(origins string) ValidationResult {
	if origins == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	origins = strings.TrimSpace(origins)
	if origins == "*" {
		return ValidationResult{Valid: true, Message: "All origins allowed"}
	}

	// Split by comma and validate each origin
	parts := strings.Split(origins, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Basic URL validation (must start with http:// or https://)
		if !strings.HasPrefix(part, "http://") && !strings.HasPrefix(part, "https://") {
			return ValidationResult{Valid: false, Message: "Origins must be * or URLs (https://example.com)"}
		}
	}

	return ValidationResult{Valid: true, Message: "Valid origins"}
}

// ValidateMethods validates HTTP methods (comma-separated)
func ValidateMethods(methods string) ValidationResult {
	if methods == "" {
		return ValidationResult{Valid: true, Message: ""}
	}

	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "OPTIONS": true, "HEAD": true, "CONNECT": true, "TRACE": true,
	}

	parts := strings.Split(methods, ",")
	for _, part := range parts {
		method := strings.ToUpper(strings.TrimSpace(part))
		if method == "" {
			continue
		}

		if !validMethods[method] {
			return ValidationResult{Valid: false, Message: "Invalid HTTP method: " + part}
		}
	}

	return ValidationResult{Valid: true, Message: "Valid methods"}
}

// GetValidationForField returns the appropriate validation result for a specific field
func GetValidationForField(templateKey, fieldName, value string) ValidationResult {
	switch templateKey {
	case "ip_restricted":
		switch fieldName {
		case "lan_subnet":
			return ValidateCIDR(value)
		case "external_ip":
			return ValidateIP(value)
		}

	case "cors_headers":
		switch fieldName {
		case "origins":
			return ValidateOrigins(value)
		case "methods":
			return ValidateMethods(value)
		}

	case "rate_limiting":
		switch fieldName {
		case "requests_per_second":
			return ValidatePositiveInt(value)
		case "burst_size":
			return ValidatePositiveInt(value)
		}

	case "large_uploads":
		if fieldName == "max_size" {
			return ValidateSize(value)
		}

	case "extended_timeouts":
		switch fieldName {
		case "read_timeout", "write_timeout", "dial_timeout":
			return ValidateDuration(value)
		}

	case "static_caching":
		if fieldName == "max_age" {
			return ValidatePositiveInt(value)
		}

	case "compression_advanced":
		if fieldName == "compression_level" {
			return ValidateIntRange(value, 1, 9)
		}

	case "https_backend":
		if fieldName == "keepalive" {
			return ValidatePositiveInt(value)
		}
	}

	return ValidationResult{Valid: true, Message: ""}
}
