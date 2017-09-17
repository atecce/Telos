package main

import (
	"flag"

	"github.com/de-nova-stella/investigations/lyrics_net"
)

func main() {

	// set start flag
	start := flag.String("s", "0", "Specify start artist of crawl.")
	flag.Parse()

	// start the investigation
	wittgenstein := lyrics_net.Investigator{
		URL: "http://www.lyrics.net",
	}
	wittgenstein.Investigate(*start)
}
