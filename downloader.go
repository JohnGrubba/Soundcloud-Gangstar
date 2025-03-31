package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Constants for readability and maintainability
const (
	outputFormat = "flac"
	audioQuality = "0"           // Best quality
	concurrentDL = "5"           // Number of concurrent fragment downloads
	illegalChars = "/\\:*?\"<>|" // Windows illegal filename characters
)

// sanitizeFilename removes illegal characters from filenames
// to ensure compatibility with the file system
func sanitizeFilename(name string) string {
	sanitized := name
	for _, char := range illegalChars {
		sanitized = strings.ReplaceAll(sanitized, string(char), "")
	}
	return sanitized
}

// buildOutputPath constructs the complete file path for the downloaded track
func buildOutputPath(filename string, playlistFileDir string) string {
	sanitized := sanitizeFilename(filename)
	// Navigate up one directory from script location and into playlist directory
	return filepath.Join("..", playlistFileDir, sanitized)
}

// getOAuthToken extracts the OAuth token from the COOKIE environment variable
// Returns empty string if not found
func getOAuthToken() (string, error) {
	cookie := os.Getenv("COOKIE")
	if cookie == "" {
		return "", fmt.Errorf("COOKIE environment variable not set")
	}

	parts := strings.Split(cookie, "oauth_token=")
	if len(parts) < 2 {
		return "", fmt.Errorf("oauth_token not found in COOKIE")
	}

	token := strings.Split(parts[1], ";")[0]
	if token == "" {
		return "", fmt.Errorf("empty oauth_token found in COOKIE")
	}

	return token, nil
}

// fileExists checks if a file exists at the given path
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// downloadFile downloads the audio file from the given URL using yt-dlp
func downloadFile(url string, outputPath string, oauthToken string) error {
	log.Printf("Downloading from %s to %s\n", url, outputPath)

	cmd := exec.Command(
		"./yt-dlp.exe",
		url,
		"-f", "ba",
		"--extract-audio",
		"-u", "oauth",
		"-p", oauthToken,
		"--audio-format", outputFormat,
		"--concurrent-fragments", concurrentDL,
		"--output", outputPath+"."+outputFormat,
	)

	log.Println("Executing command:", cmd.String())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download failed: %w\nOutput: %s", err, string(output))
	}

	log.Println("Download completed successfully")
	return nil
}

// SaveFileUsingYTDLP downloads a file from Soundcloud using yt-dlp
// Returns:
// - "exists" if the file already exists
// - "error" if there was an error during download
// - "okay" if the file was downloaded successfully
func SaveFileUsingYTDLP(filename string, url string, playlistFileDir string) string {
	outputPath := buildOutputPath(filename, playlistFileDir)
	fullPath := outputPath + "." + outputFormat

	// Check if file already exists to avoid unnecessary downloads
	if fileExists(fullPath) {
		log.Println("File already exists in library:", fullPath)
		return "exists"
	}

	// Get OAuth token for authentication
	oauthToken, err := getOAuthToken()
	if err != nil {
		log.Printf("Error getting OAuth token: %v\n", err)
		return "error"
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Error creating directory %s: %v\n", dir, err)
		return "error"
	}

	// Download the file
	if err := downloadFile(url, outputPath, oauthToken); err != nil {
		log.Printf("Error downloading file: %v\n", err)
		return "error"
	}

	return "okay"
}

// For backward compatibility with existing code
func parseFilename(filename_in string, playlistFileDir string) string {
	return buildOutputPath(filename_in, playlistFileDir)
}

// Keep original function name and signature for backward compatibility
func saveFileUsingYTDLP(filename string, url string, playlistFileDir string) string {
	return SaveFileUsingYTDLP(filename, url, playlistFileDir)
}
