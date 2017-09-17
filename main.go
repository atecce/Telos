package main

import (
	"flag"

	"github.com/de-nova-stella/investigations/lyrics_net"
)

func main() {

	start := flag.String("s", "0", "Specify start artist of crawl.")
	flag.Parse()

	wittgenstein := lyrics_net.New(*start)
	wittgenstein.Run()
}
