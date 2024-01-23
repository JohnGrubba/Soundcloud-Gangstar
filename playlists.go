package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/buger/jsonparser"
	"github.com/joho/godotenv"
)

func fetchPlaylistTracks(scurl string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//clientID := os.Getenv("CLIENT_ID")
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

	_, _ = jsonparser.ArrayEach([]byte(js), func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		track_id, err := jsonparser.GetInt(value, "id")
		if err != nil {
			fmt.Println("Error getting track id:", err)
			return
		}
		fmt.Println(track_id)

		// Fetch more track information
		// Make HTTP GET request to the final URL
		finalURL := "https://api-v2.soundcloud.com/tracks?ids=" + fmt.Sprint(track_id) + "&client_id=" + os.Getenv("CLIENT_ID")
		req, err = http.NewRequest("GET", finalURL, nil)
		if err != nil {
			fmt.Println("Error creating GET request:", err)
			return
		}
		req.Header.Set("Authorization", "OAuth "+oauthTkn)
		req.Header.Set("Cookie", cookie)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error making GET request for m3ufilegetting:", err)
			return
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		// Get the URL of the best quality song
		bestQuality, _ := jsonparser.GetString([]byte(body), "[0]", "media", "transcodings", "[0]", "url")
		format, _ := jsonparser.GetString(body, "[0]", "media", "transcodings", "[0]", "quality")
		if err != nil {
			fmt.Println("Error parsing JSON1:", err)
			return
		}
		track_auth, err := jsonparser.GetString(body, "[0]", "track_authorization")
		if err != nil {
			fmt.Println("Error parsing JSON2:", err)
			return
		}
		baseURL := bestQuality

		// Get M3U Thingy
		raw := getM3UContents(baseURL, track_auth)

		// Extract Song Details
		songTitle, err := jsonparser.GetString(body, "[0]", "title")
		if err != nil {
			fmt.Println("Error extracting Title (Strange Error)")
			return
		}
		fmt.Println("Song", songTitle, "with format", format)
		downloadFileFromM3U("songs/"+songTitle, raw)
		fmt.Println("-----------------Downloaded------------------")
	}, "[8]", "data", "tracks")

}
