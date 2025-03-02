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
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

// getAuthCredentials loads and returns the OAuth token and cookie from environment variables
// Returns oauthToken, cookie, error
func getAuthCredentials() (string, string, error) {
	if err := godotenv.Load(); err != nil {
		return "", "", fmt.Errorf("error loading .env file: %w", err)
	}

	cookie := os.Getenv("COOKIE")
	if cookie == "" {
		return "", "", fmt.Errorf("COOKIE not found in environment variables")
	}

	cookieParts := strings.Split(cookie, "oauth_token=")
	if len(cookieParts) < 2 {
		return "", "", fmt.Errorf("oauth_token not found in cookie")
	}

	oauthToken := strings.Split(cookieParts[1], ";")[0]
	return oauthToken, cookie, nil
}

// makeAuthenticatedRequest performs an authenticated HTTP request to the specified URL
// Returns response body as []byte and any error encountered
func makeAuthenticatedRequest(url string) ([]byte, error) {
	oauthToken, cookie, err := getAuthCredentials()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+oauthToken)
	req.Header.Set("Cookie", cookie)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

// extractHydrationData extracts the SoundCloud hydration data from HTML
func extractHydrationData(doc *goquery.Document) (string, error) {
	var js string
	var found bool

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		script := s.Text()
		if strings.Contains(script, "window.__sc_hydration") {
			js = strings.TrimPrefix(script, "window.__sc_hydration = ")
			js = strings.TrimSuffix(js, ";")
			found = true
		}
	})

	if !found {
		return "", fmt.Errorf("hydration data not found in HTML")
	}

	return js, nil
}

// fetchTrackIDs retrieves all track IDs from a SoundCloud playlist URL
func fetchTrackIDs(scurl string) []int64 {
	log.Printf("Fetching track IDs from: %s", scurl)

	body, err := makeAuthenticatedRequest(scurl)
	if err != nil {
		log.Printf("Error fetching playlist page: %v", err)
		return []int64{}
	}

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Error parsing HTML: %v", err)
		return []int64{}
	}

	// Extract hydration data
	js, err := extractHydrationData(doc)
	if err != nil {
		log.Printf("Error extracting hydration data: %v", err)
		return []int64{}
	}

	track_ids := make([]int64, 0)
	found := false

	jsonparser.ArrayEach([]byte(js), func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		hydratable, err := jsonparser.GetString(value, "hydratable")
		if err != nil {
			return
		}

		if hydratable == "playlist" {
			found = true
			_, err = jsonparser.ArrayEach(value, func(trackValue []byte, dataType jsonparser.ValueType, offset int, _ error) {
				track_id, err := jsonparser.GetInt(trackValue, "id")
				if err != nil {
					log.Printf("Error getting track id: %v", err)
					return
				}
				track_ids = append(track_ids, track_id)
			}, "data", "tracks")

			if err != nil {
				log.Printf("Error iterating tracks: %v", err)
			}
		}
	})

	if !found {
		log.Println("Playlist data not found in response")
		return []int64{}
	}

	// Reverse the track_ids slice to maintain original playlist order
	for i, j := 0, len(track_ids)-1; i < j; i, j = i+1, j-1 {
		track_ids[i], track_ids[j] = track_ids[j], track_ids[i]
	}

	log.Printf("Found %d tracks in playlist", len(track_ids))
	return track_ids
}

// fetchTrackInformationFromID retrieves detailed information about a track using its ID
func fetchTrackInformationFromID(track_id int64) ([]byte, error) {
	log.Printf("Fetching information for track ID: %d", track_id)

	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("CLIENT_ID not found in environment variables")
	}

	finalURL := fmt.Sprintf("https://api-v2.soundcloud.com/tracks?ids=%d&client_id=%s", track_id, clientID)
	body, err := makeAuthenticatedRequest(finalURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching track information: %w", err)
	}

	return body, nil
}

// downloadFromTrackID downloads a single track using its ID
// Returns true if download was skipped due to refresh logic
func downloadFromTrackID(track_id int64, playlistFileDir string, errored_urls *[]string, refresh bool) bool {
	body, err := fetchTrackInformationFromID(track_id)
	if err != nil {
		log.Printf("Error fetching track information: %v", err)
		*errored_urls = append(*errored_urls, fmt.Sprintf("Track ID: %d", track_id))
		return false
	}

	songTitle, err := jsonparser.GetString(body, "[0]", "title")
	if err != nil {
		log.Printf("Error extracting title: %v", err)
		*errored_urls = append(*errored_urls, fmt.Sprintf("Track ID: %d", track_id))
		return false
	}

	permalink_url, err := jsonparser.GetString(body, "[0]", "permalink_url")
	if err != nil {
		log.Printf("Error extracting permalink URL: %v", err)
		*errored_urls = append(*errored_urls, songTitle)
		return false
	}

	log.Printf("Downloading '%s' from %s", songTitle, permalink_url)
	status := saveFileUsingYTDLP(songTitle, permalink_url, playlistFileDir)

	switch status {
	case "error":
		color.Red("Error downloading '%s'", songTitle)
		*errored_urls = append(*errored_urls, songTitle)
	case "exists":
		if refresh {
			color.Green("Downloaded all new songs (Refreshed Playlist)")
			return true
		}
		color.Yellow("Track '%s' already exists", songTitle)
	case "downloaded":
		color.Blue("-----------------Downloaded------------------")
	}

	return false
}

// fetchPlaylistTracks downloads or refreshes all tracks in a SoundCloud playlist
func fetchPlaylistTracks(scurl string, playlistFileDir string, refresh bool) {
	log.Printf("Processing playlist: %s", scurl)

	track_ids := fetchTrackIDs(scurl)
	if len(track_ids) == 0 {
		color.Red("No tracks found in playlist")
		return
	}

	color.Cyan("Total Tracks: %d", len(track_ids))
	errored_urls := make([]string, 0)

	// Loop over track_ids
	for i, track_id := range track_ids {
		log.Printf("Processing track %d/%d (ID: %d)", i+1, len(track_ids), track_id)
		if downloadFromTrackID(track_id, playlistFileDir, &errored_urls, refresh) {
			break
		}
	}

	if len(errored_urls) > 0 {
		color.Red("Failed to download %d tracks:", len(errored_urls))
		for _, url := range errored_urls {
			color.Red("- %s", url)
		}
	} else {
		color.Green("All tracks processed successfully")
	}
}
