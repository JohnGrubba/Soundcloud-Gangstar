package main

import (
	"fmt"
	"os"

	"github.com/buger/jsonparser"
)

var playlists = map[string]string{
	"ColorBass": "https://soundcloud.com/jonasgrubbauer/sets/color-bass",
	"DnB":       "https://soundcloud.com/jonasgrubbauer/sets/drum-and-bass",
	"BassHouse": "https://soundcloud.com/jonasgrubbauer/sets/bass-house",
	"Dubstep":   "https://soundcloud.com/jonasgrubbauer/sets/dubstep",
	"CarSet":    "https://soundcloud.com/jonasgrubbauer/sets/goofyaahcarset",
	"Uptempo":   "https://soundcloud.com/jonasgrubbauer/sets/hardstyle-rawstyle-1",
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}

	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func main() {
	// Make CLI Program with menu
	// Display menu and ask user for action

	for {
		fmt.Println("Welcome to Gangstar CLI")
		fmt.Println("1. Refresh all Soundcloud Playlists")
		fmt.Println("2. Compare Local Playlist with Remote Playlist")
		fmt.Println("3. Download Single Track")
		fmt.Println("4. Exit")

		var input int
		fmt.Scanln(&input)

		switch input {
		case 1:
			{
				fmt.Println("Refreshing all Soundcloud Playlists")
				// Fetch all playlists
				for k, v := range playlists {
					fetchPlaylistTracks(v, k, true)
				}
			}
		case 2:
			{
				fmt.Println("Comparing Local Playlist with Remote Playlist")
				for k, v := range playlists {
					fmt.Println("Comparing Playlist", k, "with Remote Playlist")
					remote_filenames := make([]string, 0)
					track_ids := fetchTrackIDs(v)
					for _, track_id := range track_ids {
						body := fetchTrackInformationFromID(track_id)
						songTitle, _ := jsonparser.GetString(body, "[0]", "title")
						filename := parseFilename(songTitle, k)
						remote_filenames = append(remote_filenames, filename)
					}
					// Get local filenames
					local_filenames := make([]string, 0)
					files, err := os.ReadDir(k)
					if err != nil {
						fmt.Println("Error reading local directory:", err)
						return
					}
					for _, file := range files {
						local_filenames = append(local_filenames, file.Name())
					}
					fmt.Println(difference(remote_filenames, local_filenames))
				}
			}
		case 3:
			{
				fmt.Println("Downloading Single Track")
				fmt.Println("Enter Soundcloud URL")
				var url string
				fmt.Scanln(&url)
				downloadFromURL(url)
			}
		case 4:
			{
				fmt.Println("Exiting")
				return
			}
		}
	}

}
