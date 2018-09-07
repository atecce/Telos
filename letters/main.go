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

type job struct {
	client *client.APIClient
	path   string

	head string
	char string
	url  string
}

func (j job) run() error {

	log.Println("putting file at path", j.path)
	if err := j.client.PutFileURL(repoName, j.head, j.char, j.url, false, true); err != nil {
		log.Println("putting with head:", j.head)
		return err
	}
	return nil
}

func main() {

	pachyderm, err := client.NewOnUserMachine(true, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := pachyderm.CreateRepo(repoName); err != nil {
		log.Println("creating repo:", err)
	}

	var wg sync.WaitGroup
	pool := make(chan job, 100)
	go func() {
		for {
			job := <-pool
			if err := job.run(); err != nil {
				log.Println("error running job", err)
			}
			wg.Done()
		}
	}()

	commit, err := pachyderm.StartCommit(repoName, branchName)
	if err != nil {
		log.Println("starting commit:", err)
	}

	head := commit.GetID()

	u, _ := url.Parse("http://www.lyrics.net")

	for _, c := range "0ABCDEFGHIJKLMNOPQRSTUVWXYZ" {

		char := string(c)
		u.Path = path.Join("artists", char, "99999")
		url := u.String()
		pool <- job{
			client: pachyderm,
			path:   u.Path,
			head:   head,
			char:   char,
			url:    url,
		}
		wg.Add(1)
	}

	wg.Wait()

	if err := pachyderm.FinishCommit(repoName, commit.GetID()); err != nil {
		log.Println("finishing commit:", err)
	}
}
