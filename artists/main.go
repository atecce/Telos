package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

func main() {

	sem := make(chan struct{}, 1000)
	var wg sync.WaitGroup

	if walkErr := filepath.Walk("/pfs/letters/", func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		root, err := html.Parse(f)
		if err != nil {
			return err
		}

		for _, link := range scrape.FindAll(root, func(n *html.Node) bool {

			isArtist, err := regexp.MatchString("^artist/*", scrape.Attr(n, "href"))
			if err != nil {
				log.Println("matching artist link", err)
				return false
			}

			if n.Parent != nil {
				return n.Parent.Data == "strong" && n.Data == "a" && isArtist
			}

			return false
		}) {
			if link.FirstChild != nil {

				wg.Add(1)
				sem <- struct{}{}

				go func(path string) {
					defer wg.Done()

					u, _ := url.Parse("http://www.lyrics.net")

					u.Path = path
					fPath := filepath.Join("/", "pfs", "out", strings.Split(path, "/")[1])

					url := u.String()
					log.Println("GET", url)
					res, err := http.Get(url)
					if err != nil {
						log.Println("getting url:", err)
					}

					f, err := os.Create(fPath)
					if err != nil {
						log.Println("creating file at path", fPath)
					}
					defer f.Close()

					_, err = io.Copy(f, res.Body)
					if err != nil {
						log.Println("copying res", err)
					}

					<-sem

				}(scrape.Attr(link, "href"))

			}
		}

		return nil

	}); walkErr != nil {
		log.Println("walking:", walkErr)
	}

	wg.Wait()
}
