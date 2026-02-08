package validator

import (
	"net/url"
	"strings"
)

// ValidateURL validates if the URL is a valid video URL
func ValidateURL(videoURL string, allowedDomains []string) bool {
	u, err := url.Parse(videoURL)
	if err != nil {
		return false
	}

	host := u.Host
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}

	// Normalize host to lowercase for comparison
	host = strings.ToLower(host)

	for _, domain := range allowedDomains {
		// Trim whitespace and convert to lowercase
		cleanDomain := strings.ToLower(strings.TrimSpace(domain))
		if len(cleanDomain) == 0 {
			continue
		}

		// Check if host matches or contains domain
		if host == cleanDomain || strings.HasSuffix(host, "."+cleanDomain) || strings.Contains(host, cleanDomain) {
			return true
		}
	}

	return false
}

// ValidateFormatID validates format ID
func ValidateFormatID(formatID string) bool {
	if len(formatID) == 0 || len(formatID) > 50 {
		return false
	}
	return true
}

// SanitizeFilename removes dangerous characters from filename
func SanitizeFilename(filename string) string {
	dangerousChars := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*", "\x00"}
	result := filename
	for _, char := range dangerousChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// TruncateFilename truncates filename to max length while preserving extension
// Uses rune-level truncation to properly handle UTF-8 multi-byte characters
func TruncateFilename(filename string, maxLen int) string {
	// Convert to runes for proper UTF-8 handling
	runes := []rune(filename)

	// If already short enough, return as-is
	if len(runes) <= maxLen {
		return filename
	}

	// Find the extension
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		// No extension, just truncate at rune boundary
		return string(runes[:maxLen])
	}

	// Convert extension back to runes to get its rune length
	ext := filename[lastDot:]
	extRunes := []rune(ext)

	// Calculate available space for base name
	availableLen := maxLen - len(extRunes)
	if availableLen <= 0 {
		// Extension is too long, just truncate to max length at rune boundary
		return string(runes[:maxLen])
	}

	// Truncate base name at rune boundary and add extension
	baseName := string(runes[:availableLen])
	return baseName + ext
}
