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

type job struct {
	path string
}

func (j job) run() error {

	u, _ := url.Parse("http://www.lyrics.net")

	u.Path = j.path
	fPath := filepath.Join("/", "pfs", "out", strings.Split(j.path, "/")[1])

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

	_, err = io.Copy(f, res.Body)
	if err != nil {
		log.Println("copying res", err)
	}

	return nil
}

func main() {

	pool := make(chan job, 100)

	var wg sync.WaitGroup

	go func() {
		for {
			job := <-pool
			if err := job.run(); err != nil {
				log.Println("error running job", err)
			}
			wg.Done()
		}
	}()

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

				pool <- job{scrape.Attr(link, "href")}
				wg.Add(1)
			}
		}

		return nil

	}); walkErr != nil {
		log.Println("walking:", walkErr)
	}

	wg.Wait()
}
