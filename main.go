package main

import (
	"log"

	"github.com/pachyderm/pachyderm/src/client"
)

func main() {

	pachyderm, err := client.NewOnUserMachine(true, "")
	if err != nil {
		log.Fatal(err)
	}

	if err := pachyderm.CreateRepo("investigations"); err != nil {
		log.Println("creating investigations:", err)
	}

	if err := pachyderm.CreateBranch("investigations", "master", "", nil); err != nil {
		log.Println("creating master branch:", err)
	}

	commit, err := pachyderm.StartCommit("investigations", "master")
	if err != nil {
		log.Println("starting commit:", err)
	}

	if err := pachyderm.PutFileURL("investigations", commit.GetID(), "root", "http://www.lyrics.net", false, true); err != nil {
		log.Println("putting file:", err)
	}
}
