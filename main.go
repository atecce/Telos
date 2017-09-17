package main

import (
	"flag"

	"github.com/de-nova-stella/investigations/www"
)

func main() {

	start := flag.String("s", "0", "Specify start artist of crawl.")
	flag.Parse()

	wittgenstein := www.New(*start)
	wittgenstein.Run()
}
