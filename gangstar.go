package main

import (
	"fmt"
)

var playlists = map[string]string{
	"ColorBass": "https://soundcloud.com/jonasgrubbauer/sets/color-bass",
	"DnB":       "https://soundcloud.com/jonasgrubbauer/sets/drum-and-bass",
	"BassHouse": "https://soundcloud.com/jonasgrubbauer/sets/bass-house",
	"Dubstep":   "https://soundcloud.com/jonasgrubbauer/sets/dubstep",
	"CarSet":    "https://soundcloud.com/jonasgrubbauer/sets/goofyaahcarset",
	"Uptempo":   "https://soundcloud.com/jonasgrubbauer/sets/hardstyle-rawstyle-1",
	"Garage":    "https://soundcloud.com/jonasgrubbauer/sets/garage",
}

func main() {
	// Make CLI Program with menu
	// Display menu and ask user for action

	for {
		fmt.Println("Welcome to Gangstar CLI")
		fmt.Println("1. Refresh all Soundcloud Playlists")
		fmt.Println("2. Download one Track")
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
				fmt.Println("Enter the URL of the track you want to download")
				var url string
				fmt.Scanln(&url)
				oauthToken, _, err := getAuthCredentials()
				if err != nil {
					fmt.Println("Error getting auth credentials")
					return
				}
				downloadFile(url, "./custom", oauthToken)
			}
		case 4:
			{
				fmt.Println("Exiting")
				return
			}
		}
	}

}
