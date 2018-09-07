package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

const repoName = "artists"

func main() {

	u, _ := url.Parse("http://www.lyrics.net")

	pachyderm, err := client.NewInCluster()
	if err != nil {
		log.Fatal("getting client", err)
	}

	if err := pachyderm.CreateRepo(repoName); err != nil {
		log.Println("creating repo:", err)
	}

	commit, err := pachyderm.StartCommit(repoName, "master")
	if err != nil {
		log.Fatal("starting commit", err)
	}

	head := commit.GetID()

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

				if err := pachyderm.PutFileURL(repoName, head, strings.Split(u.Path, "/")[1], url, false, true); err != nil {
					log.Println("putting file with url:", url)
				}
			}
		}

		return nil

	}); walkErr != nil {
		log.Println("walking:", walkErr)
	}

	pachyderm.FinishCommit(repoName, head)
}
