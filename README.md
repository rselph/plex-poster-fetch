# plex-poster-fetch
 Download movie posters from a plex server

This is a quick little command line utility to grab jpegs of movie posters from your plex server.
It's pretty basic.

`plex-poster-fetch -plex <your-server-url> -library <movie-library> | -playlist <movie-playlist>`

If you want, you can use an environment variable to specify the plex server, so you don't have
to put it on each command line:
`export PLEX=<your-server-url>`

The files are always downloaded into the current directory.
