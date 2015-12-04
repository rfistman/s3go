package s3

import (
	"io"
	"net/http"
)

func S3ToHttpRequest(s3 *S3Request, body io.Reader) (*http.Request, error) {
	// could add it to map, but that would change this
	auth := s3.authorizationString()

	req, err := http.NewRequest(s3.Method, "http://"+s3.Header.Get("Host")+s3.URL.Path, body)
	if err != nil {
		return nil, err
	}

	req.Header = s3.Header // TODO: remove

	req.Header.Add("Authorization", auth)
	return req, nil
}

func (r *SDBRequest) HttpRequest() (*http.Request, error) {
	url := "http://" + r.host
	query := r.canonicalizedQueryString() + "&Signature=" + r.signature()
	if len(query) > 0 {
		url += "?" + query
	}

	req, err := http.NewRequest(r.httpVerb, url, nil)
	if err != nil {
		return nil, err
	}

	return req, err
}
