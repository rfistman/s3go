package util // sorry

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// so that's how you make ReadClosers
type NopCloser struct {
	io.Reader
}

// BUG: something has gone seriously wrong here
func (nc NopCloser) ReadShorts(p []int16) (int, error) {
	return ReadShorts(nc, p)
}

func (nc NopCloser) ReadFloats(p []float32) (int, error) {
	return ReadFloats(nc, p)
}

func (NopCloser) Close() error { return nil }

func LogReqAsCurl(req *http.Request) {
	verb := req.Method
	curl := "curl "
	if verb != "GET" {
		curl += fmt.Sprintf("-X %v ", verb)
	}

	curl += fmt.Sprintf("\"%v\" ", req.URL.String())

	for v, k := range req.Header {
		// No set in golang
		if v == "Accept" || v == "User-Agent" {
			continue
		}
		curl += fmt.Sprintf("-H \"%v: %v\" ", v, k[0])
	}

	if req.Body != nil {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err) // you fucked up now
		}
		err = req.Body.Close()
		if err != nil {
			panic(err)
		}

		curl += fmt.Sprintf("-d '%v'", string(buf))
		req.Body = NopCloser{bytes.NewReader(buf)}
	}

	log.Printf("%v\n", curl)
}
