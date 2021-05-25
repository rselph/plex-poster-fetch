package main

import (
	"encoding/xml"
	"flag"
	"fmt"
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
	listPlaylists  bool
	listLibraries  bool
)

func main() {
	flag.StringVar(&selectPlaylist, "playlist", "", "playlist to get images from")
	flag.StringVar(&selectLibrary, "library", "", "library to get images from")
	flag.BoolVar(&listPlaylists, "list-playlists", false, "list all playlists")
	flag.BoolVar(&listLibraries, "list-libraries", false, "list all libraries")
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

	if listLibraries {
		fetchLibraryList()
	}

	if listPlaylists {
		fetchPlaylistList()
	}

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

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatal(resp.Status)
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	err = xml.Unmarshal(bodyBytes, objOut)
	if err != nil {
		log.Fatal(err)
	}

	return
}

type Video struct {
	Title string `xml:"title,attr"`
	Year  string `xml:"year,attr"`
	Thumb string `xml:"thumb,attr"`
}

func (v *Video) fileName() string {
	name := fmt.Sprintf("%s (%s).jpg", v.Title, v.Year)
	name = strings.Map(func(r rune) rune {
		switch r {
		case ':', '/', '\\', '?', '*':
			return ' '
		default:
			return r
		}
	}, name)
	return name
}

type VideoList struct {
	Videos []*Video `xml:"Video"`
}

func fetchPosters(list *VideoList) {
	for _, v := range list.Videos {
		fetchPoster(v)
	}
}

func fetchPoster(v *Video) {
	resp, err := http.Get(srv + v.Thumb)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if len(bodyBytes) == 0 {
		return
	}

	fileName := v.fileName()
	fmt.Println(fileName)
	err = ioutil.WriteFile(fileName, bodyBytes, 0644)
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

	var onePlaylist *VideoList
	plexGet(key, &onePlaylist)
	fetchPosters(onePlaylist)
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

	var oneLibrary *VideoList
	plexGet("/library/sections/"+key+"/all", &oneLibrary)
	fetchPosters(oneLibrary)
}

func fetchLibraryList() {
	var allLibraries struct {
		Libraries []struct {
			Key   string `xml:"key,attr"`
			Title string `xml:"title,attr"`
		} `xml:"Directory"`
	}
	plexGet("/library/sections", &allLibraries)

	for _, l := range allLibraries.Libraries {
		fmt.Printf("    %s\n", l.Title)
	}
}

func fetchPlaylistList() {
	var allPlaylists struct {
		Playlists []struct {
			Key   string `xml:"key,attr"`
			Title string `xml:"title,attr"`
		} `xml:"Playlist"`
	}
	plexGet("/playlists", &allPlaylists)

	for _, p := range allPlaylists.Playlists {
		fmt.Printf("    %s\n", p.Title)
	}
}
