package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/buger/jsonparser"
	"github.com/joho/godotenv"
)

func getM3UContents(baseURL string, track_authorization string) []byte {
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

func downloadFileFromM3U(filename string, raw []byte) {
	cookie := os.Getenv("COOKIE")
	cookieParts := strings.Split(cookie, "oauth_token=")
	oauthTkn := strings.Split(cookieParts[1], ";")[0]
	// Write the music file
	f, err := os.Create(filename + ".wav")
	if err != nil {
		fmt.Println("Error creating music file:", err)
		return
	}
	defer f.Close()

	// Extract the initialization URL from the M3U file
	initURL := ""
	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		if strings.Contains(line, "#EXT-X-MAP:URI") {
			initURL = strings.ReplaceAll(line, "#EXT-X-MAP:URI=\"", "")
			initURL = strings.ReplaceAll(initURL, "\"", "")
			break
		}
	}

	// Make HTTP GET request to the initialization URL
	req, err := http.NewRequest("GET", initURL, nil)
	if err != nil {
		fmt.Println("Error creating GET request:", err)
		return
	}
	req.Header.Set("Authorization", "OAuth "+oauthTkn)
	req.Header.Set("Cookie", cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return
	}
	defer res.Body.Close()

	// Write the initialization data to the music file
	_, err = io.Copy(f, res.Body)
	if err != nil {
		fmt.Println("Error writing initialization data to music file:", err)
		return
	}

	// Download the remaining segments of the music file
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			req, err = http.NewRequest("GET", line, nil)
			if err != nil {
				fmt.Println("Error creating GET request:", err)
				return
			}
			req.Header.Set("Authorization", "OAuth "+cookie)
			req.Header.Set("Cookie", cookie)
			res, err = http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("Error making GET request:", err)
				return
			}
			defer res.Body.Close()

			// Write the segment data to the music file
			_, err = io.Copy(f, res.Body)
			if err != nil {
				fmt.Println("Error writing segment data to music file:", err)
				return
			}
		}
	}
}

func downloadFromURL(scurl string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cookie := os.Getenv("COOKIE")
	cookieParts := strings.Split(cookie, "oauth_token=")
	oauthTkn := strings.Split(cookieParts[1], ";")[0]

	// Make HTTP GET request to the initial URL
	req, err := http.NewRequest("GET", scurl, nil)
	if err != nil {
		fmt.Println("Error creating GET request:", err)
		return
	}
	req.Header.Set("Authorization", "OAuth "+oauthTkn)
	req.Header.Set("Cookie", cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request for m3ufilegetting:", err)
		return
	}
	defer res.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	// Find all script tags in the HTML
	var js string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		script := s.Text()
		if strings.Contains(script, "window.__sc_hydration") {
			js = strings.TrimPrefix(script, "window.__sc_hydration = ")
			js = strings.TrimSuffix(js, ";")
			return
		}
	})

	// Parse the JSON data from the script tag
	fmt.Println(js)

	// Get the URL of the best quality song
	bestQuality, err := jsonparser.GetString([]byte(js), "[8]", "data", "media", "transcodings", "[0]", "url")
	if err != nil {
		fmt.Println("Error parsing JSON1:", err)
		return
	}
	baseURL := bestQuality

	track_auth, err := jsonparser.GetString([]byte(js), "[8]", "data", "track_authorization")
	if err != nil {
		fmt.Println("Error parsing JSON2:", err)
		return
	}

	// Extract Song Details
	songTitle, err := jsonparser.GetString([]byte(js), "[8]", "data", "title")
	if err != nil {
		fmt.Println("Error extracting Title (Strange Error)")
		return
	}

	// GET M3U Contents
	raw := getM3UContents(baseURL, track_auth)

	downloadFileFromM3U(songTitle, raw)

	fmt.Println("Music file downloaded successfully")
}
