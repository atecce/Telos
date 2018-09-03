package main

import (
	"log"
	"net/url"
	"path"

	"github.com/pachyderm/pachyderm/src/client"
)

const (
	repoName   = "investigations"
	branchName = "master"
)

func main() {

	pachyderm, err := client.NewOnUserMachine(true, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := pachyderm.CreateRepo(repoName); err != nil {
		log.Println("creating investigations:", err)
	}

	commit, err := pachyderm.StartCommit(repoName, branchName)
	if err != nil {
		log.Println("starting commit:", err)
	}

	if err := pachyderm.CreateBranch(repoName, branchName, commit.GetID(), nil); err != nil {
		log.Println("creating master branch:", err)
	}

	branch, err := pachyderm.InspectBranch(repoName, branchName)
	if err != nil {
		log.Println("inspecting branch:", err)
	}

	head := branch.Head.GetID()

	u, _ := url.Parse("http://www.lyrics.net")

	for _, c := range "0ABCDEFGHIJKLMNOPQRSTUVWXYZ" {

		char := string(c)
		u.Path = path.Join("artists", char, "99999")

		log.Println("putting file at path", u.Path)

		if err := pachyderm.PutFileURL(repoName, head, char, u.String(), false, true); err != nil {
			log.Println("putting with head:", head)
			log.Println("err:", err)
		}
	}

	if err := pachyderm.FinishCommit(repoName, commit.GetID()); err != nil {
		log.Println("finishing commit:", err)
	}
}
