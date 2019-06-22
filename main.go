package main

import (
	"fmt"
	"log"

	"github.com/zmb3/spotify"
)

type MyClient struct {
	SpotifyClient *spotify.Client
}

func (c *MyClient) GetAllSongsOfArtist(artistId spotify.ID) []spotify.SimpleTrack {
	var tracks []spotify.SimpleTrack
	limit := 50
	offset := 0
	total := 1 // totalは初回のAPI取得で更新される
	country := spotify.CountryJapan
	for ; total > offset * limit; offset++ {
		options := spotify.Options {
			Limit: &limit,
			Offset: &offset,
			Country: &country,
		}
		partedAlbumsRes, err := c.SpotifyClient.GetArtistAlbumsOpt(artistId, &options, nil)
		if err != nil {
			log.Fatal(err)
		}
		total = partedAlbumsRes.Total
		for _, album := range partedAlbumsRes.Albums {
			tracksRes, err := c.SpotifyClient.GetAlbumTracksOpt(album.ID, 50, 0)
			if err != nil {
				log.Fatal(err)
			}
			tracks = append(tracks, tracksRes.Tracks...)
		}
	}
	return tracks
}

func (c *MyClient) GetAllFollowedArtists() []spotify.FullArtist {
	var artists []spotify.FullArtist
	lastId := ""
	for {
		partedArtistsRes, err := c.SpotifyClient.CurrentUsersFollowedArtistsOpt(50, lastId)
		if err != nil {
			log.Fatal(err)
		}
		artists = append(artists, partedArtistsRes.Artists...)
		if len(partedArtistsRes.Artists) < 50 {
			break
		}
		lastId = string(artists[len(artists)-1].ID)
	}
	return artists
}

func main() {
	var client MyClient
	client.SpotifyClient = getClient()
	// use the client to make calls that require authorization
	user, err := client.SpotifyClient.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	allFollowedArtists := client.GetAllFollowedArtists()
	for _, artist := range allFollowedArtists {
		fmt.Println(artist.Name)
	}
	artistId := allFollowedArtists[0].ID
	fmt.Println(artistId)
	allTracks := client.GetAllSongsOfArtist("0hCWVMGGQnRVfDgmhwLIxq")
	for _, track := range allTracks {
		fmt.Println(track.Name)
	}
}
