# plex-poster-fetch
 Download movie posters from a plex server

This is a quick little command line utility to grab jpegs of movie posters from your plex server.
It's pretty basic.

`plex-poster-fetch -plex <your-server-url> -token <your-token> [-library <movie-library> | -playlist <movie-playlist>]`

If you want, you can use environment variables to specify the plex server and token, so you don't have
to put them on each command line:

```
export PLEX=<your-server-url>
export PLEX_TOKEN=<your-token>
```

If you're not sure where to get a token, see [this tutorial](https://www.plexopedia.com/plex-media-server/general/plex-token/).

The files are always downloaded into the current directory.
