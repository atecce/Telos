package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/atecce/investigations/common"
	"golang.org/x/net/html"
)

func main() {

	logger := common.NewLogger()

	sem := make(chan struct{}, 1000)
	var wg sync.WaitGroup

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

									wg.Add(1)
									sem <- struct{}{}

									go func(fName, uPath string) {
										defer wg.Done()

										common.PutFile(fName, uPath, logger)

										<-sem

									}(z.Token().Data, attr.Val)

									// TODO maybe account for dorothy somehow?
								}
							}
						}
					}
				}
			}
		}
	}); err != nil {
		logger.Err(fmt.Sprintf("walking: %s\n", err.Error()))
	}
}
