package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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

		// initialize artist flag
		var artistAdded bool

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

							// add artist
							if !artistAdded {
								artistAdded = true
							}

							// album links are next token
							z.Next()
							for _, attr := range z.Token().Attr {
								if attr.Key == "href" {

									u, _ := url.Parse("http://www.lyrics.net")
									u.Path = attr.Val

									// album titles are the next token
									z.Next()

									url := u.String()
									log.Println("GET", url)
									res, err := http.Get(url)
									if err != nil {
										log.Println("getting url", err)
									}
									res.Body.Close()

									fPath := filepath.Join("/", "pfs", "out", z.Token().Data)
									f, err := os.Create(fPath)
									if err != nil {
										log.Println("creating file at path", fPath)
									}
									defer f.Close()

									_, err = io.Copy(f, res.Body)
									if err != nil {
										log.Println("copying res", err)
									}
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
