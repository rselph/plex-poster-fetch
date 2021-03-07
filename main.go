package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	srv            string
	selectPlaylist string
	selectLibrary  string
)

func main() {
	flag.StringVar(&selectPlaylist, "playlist", "", "playlist to get images from")
	flag.StringVar(&selectLibrary, "library", "", "library to get images from")
	flag.StringVar(&srv, "plex", "", "URL of plex server")
	flag.Parse()

	if srv == "" {
		srv = os.Getenv("PLEX")
	}

	if srv == "" {
		fmt.Println("must specify the plex server")
		os.Exit(1)
	}

	srv = strings.TrimRight(srv, "/")

	if selectPlaylist != "" {
		fetchPlaylist()
	}

	if selectLibrary != "" {
		fetchLibrary()
	}
}

func plexGet(key string, objOut interface{}) {
	resp, err := http.Get(srv + key)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	err = xml.Unmarshal(bodyBytes, objOut)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func fetchPoster(thumb string) {
	resp, err := http.Get(srv + thumb)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fileName := strings.ReplaceAll(thumb, "/", "_") + ".jpg"
	fileName = strings.TrimLeft(fileName, "_")
	fmt.Println(fileName)
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func fetchPlaylist() {
	var allPlaylists struct {
		Playlists []struct {
			Key   string `xml:"key,attr"`
			Title string `xml:"title,attr"`
		} `xml:"Playlist"`
	}
	plexGet("/playlists", &allPlaylists)

	var key string
	for _, p := range allPlaylists.Playlists {
		if p.Title == selectPlaylist {
			key = p.Key
		}
	}
	if key == "" {
		log.Fatal("no such playlist")
	}

	var onePlaylist struct {
		Videos []struct {
			Thumb string `xml:"thumb,attr"`
		} `xml:"Video"`
	}
	plexGet(key, &onePlaylist)

	for _, video := range onePlaylist.Videos {
		fetchPoster(video.Thumb)
	}
}

func fetchLibrary() {
	var allLibraries struct {
		Libraries []struct {
			Key   string `xml:"key,attr"`
			Title string `xml:"title,attr"`
		} `xml:"Directory"`
	}
	plexGet("/library/sections", &allLibraries)

	var key string
	for _, l := range allLibraries.Libraries {
		if l.Title == selectLibrary {
			key = l.Key
		}
	}
	if key == "" {
		log.Fatal("no such library")
	}

	var oneLibrary struct {
		Videos []struct {
			Thumb string `xml:"thumb,attr"`
		} `xml:"Video"`
	}
	plexGet("/library/sections/"+key+"/all", &oneLibrary)

	for _, video := range oneLibrary.Videos {
		fetchPoster(video.Thumb)
	}
}
