package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func parseFilename(filename_in string, playlistFileDir string) string {
	// Parse Filename (remove any illegal characters)
	filename := strings.ReplaceAll(filename_in, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, ":", "")
	filename = strings.ReplaceAll(filename, "*", "")
	filename = strings.ReplaceAll(filename, "?", "")
	filename = strings.ReplaceAll(filename, "\"", "")
	filename = strings.ReplaceAll(filename, "<", "")
	filename = strings.ReplaceAll(filename, ">", "")
	filename = strings.ReplaceAll(filename, "|", "")

	filename = playlistFileDir + "/" + filename
	return filename
}

func saveFileUsingYTDLP(filename string, url string, playlistFileDir string) string {
	cookie := os.Getenv("COOKIE")
	cookieParts := strings.Split(cookie, "oauth_token=")
	oauthTkn := strings.Split(cookieParts[1], ";")[0]

	filename = parseFilename(filename, playlistFileDir)
	// Go back one directory
	filename = "../" + filename
	// Write the music file
	// Check if the file already exists
	if _, err := os.Stat(filename + ".flac"); err == nil {
		fmt.Println("Already in Library")
		return "exists"
	}

	cmd := exec.Command("./yt-dlp.exe", url, "--add-header", "Authorization: OAuth "+oauthTkn, "-x", "--audio-format", "flac", "--audio-quality", "0", "--concurrent-fragments", "5", "--output", filename+".flac")
	outp, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Error:", err.Error())
		fmt.Println("Output:", string(outp))
		return "error"
	}

	return "okay"
	// Run youtube-dl to download the music file
}
