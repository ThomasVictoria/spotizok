package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	trackNumber        = 15
	destinationAccount = "tronconeur"
)

var (
	playlists = []struct {
		OwnerName string
		ID        spotify.ID
	}{
		{
			OwnerName: "Behry",
			ID:        "1YiWnu4c0qacl10zAvkLsw",
		},
		{
			OwnerName: "Nathan",
			ID:        "2X8ZFqsvdWKKhhYvD28Zry",
		},
		{
			OwnerName: "Thomas",
			ID:        "3umLFekph9zJikUsMbDhSw",
		},
		{
			OwnerName: "Hugo",
			ID:        "4WFkkwKlogK5gKrWPVxXXn",
		},
		{
			OwnerName: "Antoine",
			ID:        "52SXl6JF7zPTuBUO7u2bSN",
		},
	}
	ownerList      = []string{}
	finalTrackList = []spotify.ID{
		"0qteaD8tpCqauSW2touXj2", // John Holt - O.K. Fred
	}
)

func main() {
	ctx := context.Background()

	// Get result from https://developer.spotify.com/documentation/web-api
	// checking the api request response
	token := &oauth2.Token{
		AccessToken:  "",
		TokenType:    "Bearer",
		RefreshToken: "",
	}

	httpClient := spotifyauth.New().Client(ctx, token)

	client := spotify.New(httpClient)

	for _, playlistID := range playlists {

		ownerList = append(ownerList, playlistID.OwnerName)

		playlistTracks, err := client.GetPlaylistItems(ctx, playlistID.ID)

		if err != nil {
			log.Fatal("Error getting playlist tracks : ", err.Error())
		}

		if len(playlistTracks.Items) > trackNumber {
			intList := generateUniqueRandomNumbers(trackNumber, trackNumber)

			for _, n := range intList {
				finalTrackList = append(finalTrackList, playlistTracks.Items[n].Track.Track.ID)
			}
		} else {
			for _, track := range playlistTracks.Items {
				finalTrackList = append(finalTrackList, track.Track.Track.ID)
			}
		}
	}

	now := time.Now()

	playlist, err := client.CreatePlaylistForUser(
		ctx,
		destinationAccount,
		fmt.Sprintf("Spotizok %s", now.Format("January 2006")),
		fmt.Sprintf("Playlist de %s, avec les playlists de %s", now.Format("January 2006"), strings.Join(ownerList, ", ")),
		true,
		false,
	)

	if err != nil {
		log.Fatal("Error creating playlist : ", err.Error())
	}

	// Insert tracks into playlist by batch of 99
	size := 99
	var j int
	for i := 0; i < len(finalTrackList); i += size {
		j += size
		if j > len(finalTrackList) {
			j = len(finalTrackList)
		}
		client.AddTracksToPlaylist(ctx, playlist.ID, finalTrackList[i:j]...)
	}
	log.Print("Playlist created")
}

func generateUniqueRandomNumbers(n, max int) []int {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	set := make(map[int]bool)
	var result []int
	for len(set) < n {
		value := rand.Intn(max)
		if !set[value] {
			set[value] = true
			result = append(result, value)
		}
	}
	return result
}
