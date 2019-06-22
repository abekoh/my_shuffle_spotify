package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/zmb3/spotify"
)

type MyClient struct {
	UserId        string
	SpotifyClient *spotify.Client
}

func shuffle(data []spotify.ID) {
	rand.Seed(time.Now().UnixNano())
	n := len(data)
	for i := n - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		data[i], data[j] = data[j], data[i]
	}
}

func (c *MyClient) GetAllTracksOfArtist(artist spotify.ID) []spotify.SimpleTrack {
	fmt.Printf("Get All Tracks of %s\n", artist)
	var playlistTrackIds []spotify.SimpleTrack
	limit := 50
	offset := 0
	total := 1 // totalは初回のAPI取得で更新される
	country := spotify.CountryJapan
	for ; total > offset*limit; offset++ {
		options := spotify.Options{
			Limit:   &limit,
			Offset:  &offset,
			Country: &country,
		}
		partedAlbumsRes, err := c.SpotifyClient.GetArtistAlbumsOpt(artist, &options, nil)
		if err != nil {
			log.Fatal(err)
		}
		total = partedAlbumsRes.Total
		l := len(partedAlbumsRes.Albums)
		for i, album := range partedAlbumsRes.Albums {
			tracksRes, err := c.SpotifyClient.GetAlbumTracksOpt(album.ID, 50, 0)
			if err != nil {
				log.Fatal(err)
			}
			playlistTrackIds = append(playlistTrackIds, tracksRes.Tracks...)
			if i != l-1 {
				time.Sleep(time.Millisecond * 200)
			}
		}
	}
	return playlistTrackIds
}

func (c *MyClient) GetAllFollowedArtists() []spotify.FullArtist {
	fmt.Println("Get All Followed Artists")
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

func (c *MyClient) GetAllPlaylists() []spotify.SimplePlaylist {
	fmt.Println("Get All Playlist")
	var playlists []spotify.SimplePlaylist
	limit := 50
	offset := 0
	total := 1 // totalは初回のAPI取得で更新される
	for ; total > offset*limit; offset++ {
		options := spotify.Options{
			Limit:  &limit,
			Offset: &offset,
		}
		partedPlaylistsRes, err := c.SpotifyClient.GetPlaylistsForUserOpt(c.UserId, &options)
		if err != nil {
			log.Fatal(err)
		}
		total = partedPlaylistsRes.Total
		playlists = append(playlists, partedPlaylistsRes.Playlists...)
	}
	return playlists
}

func (c *MyClient) PreparePlaylist(playlistName string) spotify.ID {
	fmt.Println("Prepare playlist")
	allPlaylists := c.GetAllPlaylists()
	isExisted := false
	var targetPlaylistId spotify.ID
	for _, playlist := range allPlaylists {
		fmt.Println(playlist.Name)
		if playlist.Name == playlistName {
			isExisted = true
			targetPlaylistId = playlist.ID
			break
		}
	}
	if isExisted {
		fmt.Printf("Same name playlist is found(%s). Remove tracks.\n", targetPlaylistId)
		allTracks, err := c.SpotifyClient.GetPlaylistTracks(targetPlaylistId)
		if err != nil {
			log.Fatal(err)
		}
		var trackIds []spotify.ID
		for _, track := range allTracks.Tracks {
			trackIds = append(trackIds, track.Track.ID)
		}
		_, err = c.SpotifyClient.RemoveTracksFromPlaylist(targetPlaylistId, trackIds...)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		targetPlaylist, err := c.SpotifyClient.CreatePlaylistForUser(c.UserId, playlistName, "My shuffled playlist", true)
		if err != nil {
			log.Fatal(err)
		}
		targetPlaylistId = targetPlaylist.ID
	}
	return targetPlaylistId
}

func (c *MyClient) CreateShuffledPlaylist(playlistName string) error {
	targetPlaylistId := c.PreparePlaylist(playlistName)
	allFollowedArtists := c.GetAllFollowedArtists()
	var allFollowedArtistIds []spotify.ID
	for _, artist := range allFollowedArtists {
		allFollowedArtistIds = append(allFollowedArtistIds, artist.ID)
	}
	shuffle(allFollowedArtistIds)
	artistNum := 10
	var playlistTrackIds []spotify.ID
	for _, artist := range allFollowedArtistIds[:artistNum] {
		allTracks := c.GetAllTracksOfArtist(artist)
		var allTrackIds []spotify.ID
		for _, track := range allTracks {
			allTrackIds = append(allTrackIds, track.ID)
		}
		shuffle(allTrackIds)
		playlistTrackIds = append(playlistTrackIds, allTrackIds[0:1]...)
	}
	_, err := c.SpotifyClient.AddTracksToPlaylist(targetPlaylistId, playlistTrackIds...)
	return err
}

func main() {
	var client MyClient
	client.SpotifyClient = getClient()
	user, err := client.SpotifyClient.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)
	client.UserId = user.ID

	client.CreateShuffledPlaylist("Shuffled")
}
