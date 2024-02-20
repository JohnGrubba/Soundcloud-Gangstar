package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func getSongContent(baseURL string, track_authorization string) []byte {
	cookie := os.Getenv("COOKIE")
	cookieParts := strings.Split(cookie, "oauth_token=")
	oauthTkn := strings.Split(cookieParts[1], ";")[0]

	// Construct the final URL with client ID and track authorization
	finalURL := baseURL + "?client_id=" + url.QueryEscape(os.Getenv("CLIENT_ID")) + "&track_authorization=" + url.QueryEscape(track_authorization)

	// Make HTTP GET request to the final URL
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		fmt.Println("Error creating GET request:", err)
		return nil
	}
	req.Header.Set("Authorization", "OAuth "+oauthTkn)
	req.Header.Set("Cookie", cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request for m3ufilegetting:", err)
		return nil
	}
	defer res.Body.Close()

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	// Parse the JSON response
	var m3uURL map[string]string
	err = json.Unmarshal(body, &m3uURL)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	// Get the URL of the M3U file
	m3uURLString := m3uURL["url"]
	// Make HTTP GET request to the M3U URL
	res, err = http.Get(m3uURLString)
	if err != nil {
		fmt.Println("Error making GET request form3ufile:", err)
		return nil
	}
	defer res.Body.Close()

	// Read the M3U file
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading M3U file:", err)
		return nil
	}
	return raw
}

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

	filename = playlistFileDir + "/" + filename + ".wav"
	return filename
}

func saveFileFromRAWData(filename string, raw []byte, playlistFileDir string) string {
	filename = parseFilename(filename, playlistFileDir)
	// Write the music file
	// Check if the file already exists
	if _, err := os.Stat(filename); err == nil {
		fmt.Println("Already in Library")
		return "exists"
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating music file:", err)
		return "error"
	}
	defer f.Close()

	// Write raw into file
	_, err = f.Write(raw)
	if err != nil {
		fmt.Println("Error writing music file:", err)
		return "error"
	}

	return filename + ".wav"
}
