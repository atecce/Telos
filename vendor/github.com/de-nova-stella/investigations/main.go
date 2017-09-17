package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"

	"github.com/atecce/investigations/lyrics_net"
)

func main() {

	// set up memory profile
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			f, err := os.Create("mem.prof")
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			f.Close()
		}
	}()

	// set start flag
	start := flag.String("s", "0", "Specify start artist of crawl.")
	flag.Parse()

	// start the investigation
	wittgenstein := lyrics_net.Investigator{
		URL: "http://www.lyrics.net",
	}
	wittgenstein.Investigate(*start)
}
