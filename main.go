package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	youtubeChannelID   = "UCpG0uQQzJv6l3MIWOjb1g_g"
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
		{
			OwnerName: "Solène",
			ID:        "1X4bpOl12hPzQHhlJ7aQ0e",
		},
		{
			OwnerName: "Max",
			ID:        "4ZTBr9ZYif08CvFONnECdP",
		},
		{
			OwnerName: "Tita",
			ID:        "2oAnAAHBDR6Ju3Bzu6tbDh",
		},
	}
	ownerList        = []string{}
	finalTrackListID = []spotify.ID{
		"0qteaD8tpCqauSW2touXj2",
	}
	finalTrackListName = []string{
		"John Holt - O.K. Fred",
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
				var artists string
				for _, artist := range playlistTracks.Items[n].Track.Track.Artists {
					artists += " " + artist.Name
				}

				finalTrackListID = append(
					finalTrackListID,
					playlistTracks.Items[n].Track.Track.ID,
				)
				finalTrackListName = append(
					finalTrackListName,
					playlistTracks.Items[n].Track.Track.Name+artists,
				)
			}
		} else {
			for _, trackToInsert := range playlistTracks.Items {
				var artists string
				for _, artist := range trackToInsert.Track.Track.Artists {
					artists += " " + artist.Name
				}

				finalTrackListID = append(
					finalTrackListID,
					trackToInsert.Track.Track.ID,
				)
				finalTrackListName = append(
					finalTrackListName,
					trackToInsert.Track.Track.Name+artists,
				)
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
	for i := 0; i < len(finalTrackListID); i += size {
		j += size
		if j > len(finalTrackListID) {
			j = len(finalTrackListID)
		}
		client.AddTracksToPlaylist(ctx, playlist.ID, finalTrackListID[i:j]...)
	}
	log.Print("Playlist created on spotify")

	// If need to import already existing spotify playlist
	// finalTrackListName = transferAlreadyExist(client)

	convertToYoutube(finalTrackListName)
}

func transferAlreadyExist(client *spotify.Client) []string {
	tracks := []string{}

	ctx := context.Background()

	// Spotify Playlist ID to import
	playlistID := spotify.ID("")

	playlistTracks, err := client.GetPlaylistItems(ctx, playlistID)

	if err != nil {
		log.Fatal("Error getting playlist tracks : ", err.Error())
	}

	log.Printf("Playlist has %d total tracks", playlistTracks.Total)
	for page := 1; ; page++ {
		log.Printf("  Page %d has %d tracks", page, len(playlistTracks.Items))
		for _, track := range playlistTracks.Items {

			var artists string
			for _, artist := range track.Track.Track.Artists {
				artists += " " + artist.Name
			}
			tracks = append(
				tracks,
				track.Track.Track.Name+artists,
			)
		}
		err = client.NextPage(ctx, playlistTracks)

		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	return tracks
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

func RemoveIndex(s []string, index int) []string {
	ret := make([]string, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func convertToYoutube(spotifyTrackList []string) {
	ctx := context.Background()

	// https://developers.google.com/oauthplayground
	token := &oauth2.Token{
		AccessToken:  "",
		RefreshToken: "",
		TokenType:    "Bearer",
	}

	config := oauth2.Config{}
	client := config.Client(ctx, token)

	youtubeService, err := youtube.NewService(ctx, option.WithHTTPClient(client), option.WithScopes(youtube.YoutubeScope))

	if err != nil {
		log.Print(err)
	}

	now := time.Now()

	insertedPlaylist, err := youtubeService.Playlists.Insert([]string{"snippet"}, &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			ChannelId: youtubeChannelID,
			Title:     fmt.Sprintf("Spotizok %s", now.Format("January 2006")),
		},
	}).Do()

	if err != nil {
		log.Print(err)
	}

	maxResults := flag.Int64("max-results", 1, "Max YouTube results")

	//  ====================
	// 	No forget spotifyTrackList[65:] pour finir l'insertion
	// 	====================
	// ET LA PLAYLIST ID
	// playlistID := ""
	// 	====================

	for _, track := range spotifyTrackList[65:] {
		time.Sleep(5 * time.Second)

		response, err := youtubeService.Search.List([]string{"id", "snippet"}).Q(track).MaxResults(*maxResults).Do()

		if err != nil {
			log.Print(err)
		}

		ytPlaylistItems := youtube.PlaylistItem{
			Snippet: &youtube.PlaylistItemSnippet{
				// 	====================
				// NO FORGET TO SWITCH FOR SECOND INSERT
				// 	====================
				// PlaylistId: plalistID,
				// 	====================
				PlaylistId: insertedPlaylist.Id,
				// 	====================
				ResourceId: response.Items[0].Id,
			},
		}

		_, err = youtubeService.PlaylistItems.Insert([]string{"snippet"}, &ytPlaylistItems).Do()

		if err != nil {
			log.Print(err)
		}

		log.Print(response.Items[0].Snippet.Title + " inserted")
	}

	log.Print("Playlist transféré sur Youtube")
}
