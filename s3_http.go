package s3

import "net/http"

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
