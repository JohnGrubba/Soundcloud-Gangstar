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

func downloadFromURL(sc_original_url string) int64 {
	// Fetch the track ID from the Soundcloud URL
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//clientID := os.Getenv("CLIENT_ID")
	cookie := os.Getenv("COOKIE")
	cookieParts := strings.Split(cookie, "oauth_token=")
	oauthTkn := strings.Split(cookieParts[1], ";")[0]
	// Make HTTP GET request to the initial URL
	req, err := http.NewRequest("GET", sc_original_url, nil)
	if err != nil {
		fmt.Println("Error creating GET request:", err)
		return 0
	}
	req.Header.Set("Authorization", "OAuth "+oauthTkn)
	req.Header.Set("Cookie", cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request for m3ufilegetting:", err)
		return 0
	}
	defer res.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return 0
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
	dl_url_base, err := jsonparser.GetString([]byte(js), "[8]", "data", "media", "transcodings", "[1]", "url")
	if err != nil {
		fmt.Println("Error getting track id:", err)
		return 0
	}
	track_auth, err := jsonparser.GetString([]byte(js), "[8]", "data", "track_authorization")
	if err != nil {
		fmt.Println("Error parsing JSON2:", err)
		return 0
	}
	title, err := jsonparser.GetString([]byte(js), "[8]", "data", "title")
	if err != nil {
		fmt.Println("Error parsing JSON2:", err)
		return 0
	}
	// Get M3U Thingy
	raw := getSongContent(dl_url_base, track_auth)
	saveFileFromRAWData(title, raw, ".")
	return 0

}

func fetchTrackIDs(scurl string) []int64 {
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
		return []int64{}
	}
	req.Header.Set("Authorization", "OAuth "+oauthTkn)
	req.Header.Set("Cookie", cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request for m3ufilegetting:", err)
		return []int64{}
	}
	defer res.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return []int64{}
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

	track_ids := make([]int64, 0)

	_, _ = jsonparser.ArrayEach([]byte(js), func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		track_id, err := jsonparser.GetInt(value, "id")
		if err != nil {
			fmt.Println("Error getting track id:", err)
			return
		}
		track_ids = append(track_ids, track_id)
	}, "[8]", "data", "tracks")

	// Reverse the track_ids slice
	for i, j := 0, len(track_ids)-1; i < j; i, j = i+1, j-1 {
		track_ids[i], track_ids[j] = track_ids[j], track_ids[i]
	}
	return track_ids
}

func fetchTrackInformationFromID(track_id int64) []byte {
	cookie := os.Getenv("COOKIE")
	cookieParts := strings.Split(cookie, "oauth_token=")
	oauthTkn := strings.Split(cookieParts[1], ";")[0]
	// Fetch more track information
	// Make HTTP GET request to the final URL
	finalURL := "https://api-v2.soundcloud.com/tracks?ids=" + fmt.Sprint(track_id) + "&client_id=" + os.Getenv("CLIENT_ID")
	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		fmt.Println("Error creating GET request:", err)
		return []byte{}
	}
	req.Header.Set("Authorization", "OAuth "+oauthTkn)
	req.Header.Set("Cookie", cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request for m3ufilegetting:", err)
		return []byte{}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return []byte{}
	}
	return body
}

func downloadFromTrackID(track_id int64, playlistFileDir string, errored_urls *[]string, refresh bool) bool {
	body := fetchTrackInformationFromID(track_id)
	songTitle, err := jsonparser.GetString(body, "[0]", "title")
	if err != nil {
		fmt.Println("Error extracting Title (Strange Error)")
		*errored_urls = append(*errored_urls, songTitle)
		return false
	}
	// Get the URL of the best quality song
	bestQuality, err1 := jsonparser.GetString([]byte(body), "[0]", "media", "transcodings", "[1]", "url")
	format, err := jsonparser.GetString(body, "[0]", "media", "transcodings", "[1]", "quality")
	if err != nil || err1 != nil {
		fmt.Println("Error parsing JSON:", err)
		*errored_urls = append(*errored_urls, songTitle)
		return false
	}
	// Extract Song Details
	color.Blue("Song " + songTitle + " with format " + format)
	if format != "hq" {
		fmt.Println("Skipping non-HQ song")
		*errored_urls = append(*errored_urls, songTitle)
		return false
	}
	track_auth, err := jsonparser.GetString(body, "[0]", "track_authorization")
	if err != nil {
		fmt.Println("Error parsing JSON2:", err)
		*errored_urls = append(*errored_urls, songTitle)
		return false
	}
	baseURL := bestQuality

	// Get M3U Thingy
	raw := getSongContent(baseURL, track_auth)

	// fmt.Println(string(body))
	status := saveFileFromRAWData(songTitle, raw, playlistFileDir)
	if status == "error" {
		color.Red("Error downloading song")
		// Append to errored_urls
		*errored_urls = append(*errored_urls, songTitle)
	}
	if status == "exists" && refresh {
		color.Green("Downloaded all new Songs (Refreshed Playlist)")
		return true
	}
	_ = errored_urls // Fix for go-staticcheck SA4010
	color.Blue("-----------------Downloaded------------------")
	return false
}

// Downloads or refreshes the playlist
func fetchPlaylistTracks(scurl string, playlistFileDir string, refresh bool) {
	track_ids := fetchTrackIDs(scurl)

	errored_urls := make([]string, 0)

	// Loop over track_ids
	for _, track_id := range track_ids {
		if downloadFromTrackID(track_id, playlistFileDir, &errored_urls, refresh) {
			break
		}
	}
	if len(errored_urls) > 0 {
		color.Red("Errored URLs:")
		for _, url := range errored_urls {
			color.Red(url)
		}
	}
}
