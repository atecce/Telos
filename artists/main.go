package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

func main() {

	if walkErr := filepath.Walk("/pfs/letters", func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		log.Println("opening file at", path)

		f, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}

		root, parseErr := html.Parse(f)
		if err != nil {
			return parseErr
		}

		for _, link := range scrape.FindAll(root, func(n *html.Node) bool {
			if n.Parent != nil {
				return n.Parent.Data == "strong" && n.Data == "a"
			}
			return false
		}) {
			if link.FirstChild != nil {
				log.Println(scrape.Attr(link, "href"))
				log.Println(link.FirstChild.Data)
			}
		}

		return nil

	}); walkErr != nil {
		log.Println("walking:", walkErr)
	}
}
