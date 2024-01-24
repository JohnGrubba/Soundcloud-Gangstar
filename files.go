package main

import (
	"fmt"

	"github.com/bogem/id3v2/v2"
)

func writeFileTags(fileName string, artist string, title string, track_id int64, cover []byte) {
	tag, err := id3v2.Open(fileName, id3v2.Options{Parse: true})
	if err != nil {
		fmt.Println("Error while opening file: ", err)
	}
	defer tag.Close()

	fmt.Println(artist, title, track_id)
	// Set tags
	tag.SetArtist(artist)
	tag.SetTitle(title)
	tag.AddAttachedPicture(id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     []byte(cover),
	})

	ufid_frame := id3v2.UFIDFrame{
		OwnerIdentifier: fmt.Sprint(track_id),
		Identifier:      []byte(fmt.Sprint(track_id)),
	}
	tag.AddUFIDFrame(ufid_frame)

	// Write tag to file.mp3
	if err = tag.Save(); err != nil {
		fmt.Println("Error while saving a tag: ", err)
	}
	fmt.Println("Wrote Metadata")
}
