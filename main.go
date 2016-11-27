package main

import (
	"flag"

	"github.com/atecce/investigations/lyricsdotnet"
)

func main() {

	// set start flag
	start := flag.String("s", "0", "Specify start artist of crawl.")
	flag.Parse()

	// start the investigation
	lyricsdotnet.Investigate(*start)
}
