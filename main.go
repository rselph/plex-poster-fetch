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
	"path"
)

var (
	srv            = "http://192.168.3.12:32400"
	selectPlaylist string
)

func main() {
	flag.StringVar(&selectPlaylist, "playlist", "Movie Night",
		"playlist to get images from")
	flag.Parse()
	if flag.NArg() > 0 {
		srv = flag.Arg(0)
	}

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
		resp, err := http.Get(srv + video.Thumb)
		if err != nil {
			log.Fatal(err)
		}

		fileName := path.Base(video.Thumb) + ".jpg"
		fmt.Println(fileName)
		f, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		f.Close()
		resp.Body.Close()
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
