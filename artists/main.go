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

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

func main() {

	u, _ := url.Parse("http://www.lyrics.net")

	if walkErr := filepath.Walk("/pfs/letters/", func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		f, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}

		root, parseErr := html.Parse(f)
		if err != nil {
			return parseErr
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

				u.Path = scrape.Attr(link, "href")

				url := u.String()

				res, err := http.Get(url)
				if err != nil {
					log.Println("getting url:", err)
				}

				outPath := filepath.Join("/", "pfs", "out", strings.Split(u.Path, "/")[1])

				f, err := os.Create(outPath)
				if err != nil {
					log.Println("creating file at path", outPath)
				}

				_, err = io.Copy(f, res.Body)
				if err != nil {
					log.Println("copying res", err)
				}
			}
		}

		return nil

	}); walkErr != nil {
		log.Println("walking:", walkErr)
	}
}
