package rest

import (
	"io"
	"log"
	"net/http"
	"time"
)

// gets an http resource by name that you can read from
func Get(name string) (*io.ReadCloser, bool) {

	// never stop trying
	for {
		resp, err := http.Get(name)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}

		// TODO
		log.Println(name, resp.Status)

		switch resp.StatusCode {

		case http.StatusOK:
			return &resp.Body, true
		case http.StatusForbidden:
			resp.Body.Close()
			return nil, false
		case http.StatusNotFound:
			resp.Body.Close()
			return nil, false

		case http.StatusServiceUnavailable:
			time.Sleep(10 * time.Minute)
		case http.StatusGatewayTimeout:
			time.Sleep(time.Minute)
		default:
			time.Sleep(time.Minute)
		}
	}
}
