package s3

import (
	"net/http"
)

func S3ToHttpRequest(s3 *S3Request) (*http.Request, error) {
	// could add it to map, but that would change this
	auth := s3.AuthorizationString()

	req, err := http.NewRequest(s3.httpVerb, "http://"+s3.args["Host"]+s3.resource, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range s3.args {
		req.Header.Add(k, v)
	}
	req.Header.Add("Authorization", auth)
	return req, nil
}
