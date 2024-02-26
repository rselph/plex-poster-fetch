package main

import (
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	srv            string
	token          string
	selectPlaylist string
	selectLibrary  string
	listPlaylists  bool
	listLibraries  bool
	unsafe         bool
	fanArt         bool
	debug          string
	quiet          bool

	debugRegex *regexp.Regexp
	now        = time.Now().Unix()
)

func main() {
	flag.StringVar(&selectPlaylist, "playlist", "", "playlist to get images from")
	flag.StringVar(&selectLibrary, "library", "", "library to get images from")
	flag.BoolVar(&listPlaylists, "list-playlists", false, "list all playlists")
	flag.BoolVar(&listLibraries, "list-libraries", false, "list all libraries")
	flag.BoolVar(&unsafe, "unsafe", false, "ignore certificate errors")
	flag.BoolVar(&fanArt, "fanart", false, "get fanart instead of posters")
	flag.StringVar(&srv, "plex", "", "URL of plex server")
	flag.StringVar(&token, "token", "",
		"Plex token. See https://www.plexopedia.com/plex-media-server/general/plex-token/")
	flag.BoolVar(&quiet, "q", false, "suppress non-error output")
	flag.StringVar(&debug, "debug", "", "regular expression for keys to print")
	flag.Parse()

	if srv == "" {
		srv = os.Getenv("PLEX")
	}

	if srv == "" {
		fmt.Println("must specify the plex server")
		os.Exit(1)
	}

	if token == "" {
		token = os.Getenv("PLEX_TOKEN")
	}

	srv = strings.TrimRight(srv, "/")

	if unsafe {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if debug != "" {
		debugRegex = regexp.MustCompile(debug)
	}

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

func plexGet(key string, objOut interface{}) []byte {
	resp, err := http.Get(srv + key + "?X-Plex-Token=" + token)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatal(resp.Status)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	if debugRegex != nil && debugRegex.MatchString(key) {
		fmt.Println(string(bodyBytes))
	}

	if objOut != nil {
		err = xml.Unmarshal(bodyBytes, objOut)
		if err != nil {
			log.Fatal(err)
		}
	}

	return bodyBytes
}

type Video struct {
	Title     string `xml:"title,attr"`
	Year      string `xml:"year,attr"`
	Thumb     string `xml:"thumb,attr"`
	Art       string `xml:"art,attr"`
	AddedAt   int64  `xml:"addedAt,attr"`
	UpdatedAt int64  `xml:"updatedAt,attr"`
}

func (v *Video) fileName(suffix string) string {
	name := fmt.Sprintf("%s (%s)%s.jpg", v.Title, v.Year, suffix)
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
	v.Validate()

	bodyBytes := plexGet(v.Thumb, nil)

	if len(bodyBytes) == 0 {
		return
	}

	fileName := v.fileName(" poster")
	if !quiet {
		fmt.Println(fileName)
	}
	err := os.WriteFile(fileName, bodyBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func fetchFanarts(list *VideoList) {
	for _, v := range list.Videos {
		fetchFanart(v)
	}
}

func fetchFanart(v *Video) {
	v.Validate()

	bodyBytes := plexGet(v.Art, nil)

	if len(bodyBytes) == 0 {
		return
	}

	fileName := v.fileName(" fanart")
	if !quiet {
		fmt.Println(fileName)
	}
	err := os.WriteFile(fileName, bodyBytes, 0644)
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
	if fanArt {
		fetchFanarts(onePlaylist)
	} else {
		fetchPosters(onePlaylist)
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

	var oneLibrary *VideoList
	plexGet("/library/sections/"+key+"/all", &oneLibrary)
	if fanArt {
		fetchFanarts(oneLibrary)
	} else {
		fetchPosters(oneLibrary)
	}
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

func (v *Video) Validate() {
	if v.AddedAt > now || v.UpdatedAt > now {
		fmt.Printf("WARNING: FUTURE DATE %s (%s)\n", v.Title, v.Year)
		fmt.Printf("\tAdded: %s, Updated: %s\n",
			time.Unix(v.AddedAt, 0).Format(time.DateTime),
			time.Unix(v.UpdatedAt, 0).Format(time.DateTime))
	}
}
