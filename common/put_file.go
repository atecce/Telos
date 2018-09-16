package common

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// NewLogger dials syslog and returns the logger
func NewLogger() *syslog.Writer {

	logger, err := syslog.Dial("tcp", "35.237.191.105:514", syslog.LOG_INFO, "investigations")
	if err != nil {
		log.Fatal("dialing syslog", err)
	}

	return logger
}

// PutFile takes a file name and a url path and copies the http response body to the relevant pfs file
// TODO create an object which simply owns the logger instead of passing it around
func PutFile(fName string, uPath string, logger *syslog.Writer) {

	u, _ := url.Parse("http://www.lyrics.net")
	u.Path = uPath

	fPath := filepath.Join("/", "pfs", "out", fName)

	url := u.String()
	logger.Info(fmt.Sprintf("GET %s\n", url))
	res, err := http.Get(url)
	if err != nil {
		logger.Err(fmt.Sprintf("getting url: %s\n", err.Error()))
	}
	defer res.Body.Close()

	f, err := os.Create(fPath)
	if err != nil {
		logger.Err(fmt.Sprintf("creating file at path: %s\n", fPath))
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		logger.Err(fmt.Sprintf("copying res: %s\n", err.Error()))
	}
}
