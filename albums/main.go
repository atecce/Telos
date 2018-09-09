package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/atecce/investigations/common"
	"golang.org/x/net/html"
)

func main() {

	if err := filepath.Walk("/pfs/artists/", func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// parse page
		z := html.NewTokenizer(f)
		for {
			switch z.Next() {

			// end of html document
			case html.ErrorToken:
				return nil

			// catch start tags
			case html.StartTagToken:

				// set token
				t := z.Token()

				// look for artist album labels
				if t.Data == "h3" {
					for _, a := range t.Attr {
						if a.Key == "class" && a.Val == "artist-album-label" {

							// album links are next token
							z.Next()
							for _, attr := range z.Token().Attr {
								if attr.Key == "href" {

									// album titles are the next token
									z.Next()

									common.PutFile(z.Token().Data, attr.Val)

									// TODO maybe account for dorothy somehow?
								}
							}
						}
					}
				}
			}
		}
	}); err != nil {
		log.Println("walking", err)
	}
}
