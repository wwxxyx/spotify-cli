package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zmb3/spotify"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

var redirectURI = "http://localhost:8888/spotify-cli"

var (
	auth = spotify.NewAuthenticator(
		redirectURI,
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserLibraryRead,
	)
	ch    = make(chan *spotify.Client)
	state = uuid.New().String()
)

var clientId = os.Getenv("SPOTIFY_CLIENT_ID")
var clientSecret = os.Getenv("SPOTIFY_SECRET")

func authenticate() *spotify.Client {
	auth.SetAuthInfo(clientId, clientSecret)
	url := auth.AuthURL(state)

	http.HandleFunc("/spotify-cli", authCallback)

	go http.ListenAndServe(":8888", nil)

	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		fmt.Errorf("Sorry, not supported")
	}

	if err != nil {
		log.Fatal(err)
	}

	client := <-ch
	return client
}

func authCallback(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusNotFound)
		return
	}
	client := auth.NewClient(token)
	ch <- &client

	user, err := client.CurrentUser()
	fmt.Fprintf(w, "<h1>Logged into spotify cli as:</h1>\n<p>%v</p>", user.DisplayName)
}
