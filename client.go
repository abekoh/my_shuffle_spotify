package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const redirectURI = "http://localhost:8080/callback"
const keysPath = "./keys.json"
const tokenPath = "./token.json"

var (
	scopes    = []string{spotify.ScopePlaylistModifyPrivate, spotify.ScopePlaylistModifyPublic, spotify.ScopeUserFollowRead}
	authorize = spotify.NewAuthenticator(redirectURI, scopes...)
	tokenCh   = make(chan *oauth2.Token)
	state     = "abc123"
)

type Keys struct {
	Id     string `json:"spotify_id"`
	Secret string `json:"spotify_secret"`
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	tok, err := authorize.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	fmt.Fprintf(w, "Login Completed!")
	tokenCh <- tok
}

func loginWithBrowser() *oauth2.Token {
	http.HandleFunc("/callback", authHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)
	url := authorize.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	tok := <-tokenCh
	return tok
}

func getTokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot open %s", tokenPath))
	}
	readAll, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read %s", tokenPath))
	}
	var tok *oauth2.Token
	err = json.Unmarshal(readAll, &tok)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot parse %s", tokenPath))
	}
	return tok, nil
}

func saveTokenFile(tok *oauth2.Token) error {
	tokJson, err := json.MarshalIndent(tok, "", "    ")
	if err != nil {
		return errors.New(fmt.Sprintf("cannot open %s", tokenPath))
	}
	err = ioutil.WriteFile(tokenPath, tokJson, 0666)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot write %s", tokenPath))
	}
	return nil
}

func getKeys() *Keys {
	f, err := os.Open(keysPath)
	if err != nil {
		log.Fatal(err)
	}
	readAll, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	var keys *Keys
	err = json.Unmarshal(readAll, &keys)
	if err != nil {
		log.Fatal(err)
	}
	return keys
}

func getClient() *spotify.Client {
	keys := getKeys()
	authorize.SetAuthInfo(keys.Id, keys.Secret)
	var client spotify.Client
	var tok *oauth2.Token
	tok, err := getTokenFromFile()
	if err != nil {
		tok = loginWithBrowser()
		saveTokenFile(tok)
	}
	client = authorize.NewClient(tok)
	return &client
}
