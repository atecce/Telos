package main

import (
	"log"
	"net/url"
	"path"
	"sync"

	"github.com/pachyderm/pachyderm/src/client"
)

const (
	repoName   = "letters"
	branchName = "master"
)

func main() {

	var wg sync.WaitGroup

	pachyderm, err := client.NewOnUserMachine(true, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := pachyderm.CreateRepo(repoName); err != nil {
		log.Println("creating repo:", err)
	}

	commit, err := pachyderm.StartCommit(repoName, branchName)
	if err != nil {
		log.Println("starting commit:", err)
	}

	head := commit.GetID()

	u, _ := url.Parse("http://www.lyrics.net")

	for _, c := range "0ABCDEFGHIJKLMNOPQRSTUVWXYZ" {

		fileName := string(c)
		u.Path = path.Join("artists", fileName, "99999")
		url := u.String()

		go func(fileName, path, url string) {
			defer wg.Done()

			log.Println("putting file at path", path)
			if err := pachyderm.PutFileURL(repoName, head, fileName, url, false, true); err != nil {
				log.Println("putting with head:", head)
			}
		}(fileName, u.Path, url)

		wg.Add(1)
	}
	wg.Wait()

	if err := pachyderm.FinishCommit(repoName, commit.GetID()); err != nil {
		log.Println("finishing commit:", err)
	}
}
