package common

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// PutFile takes a file name and a url path and copies the http response body to the relevant pfs file
func PutFile(fName string, uPath string) {

	u, _ := url.Parse("http://www.lyrics.net")
	u.Path = uPath

	fPath := filepath.Join("/", "pfs", "out", fName)

	url := u.String()
	log.Println("GET", url)
	res, err := http.Get(url)
	if err != nil {
		log.Println("getting url:", err)
	}
	defer res.Body.Close()

	f, err := os.Create(fPath)
	if err != nil {
		log.Println("creating file at path", fPath)
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		log.Println("copying res", err)
	}
}
