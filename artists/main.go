package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/atecce/investigations/common"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

func main() {

	logger := common.NewLogger()

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
		defer f.Close()

		root, err := html.Parse(f)
		if err != nil {
			return err
		}

		for _, link := range scrape.FindAll(root, func(n *html.Node) bool {

			isArtist, err := regexp.MatchString("^artist/*", scrape.Attr(n, "href"))
			if err != nil {
				logger.Err(fmt.Sprintf("matching artist link: %s\n", err.Error()))
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

					common.PutFile(strings.Split(path, "/")[1], path, logger)

					<-sem

				}(scrape.Attr(link, "href"))

			}
		}

		return nil

	}); walkErr != nil {
		logger.Info(fmt.Sprintf("walking: %s\n", walkErr.Error()))
	}

	wg.Wait()
}
