package s3

import (
	//	"log"
	"net/url"
	"strings"
)

type AWSRequest interface {
	StringToSign() string
	AddCredentials(cred *SecurityCredentials)
}

type Strmap map[string]string

type SDBRequest struct {
	httpVerb string
	// uri      string	// hardcoding as "/"
	query Strmap
}

func percent_encode(s string) string {
	// go's "+" encoding for space doesn't seem to please
	//return url.QueryEscape(s)
	return strings.Replace(url.QueryEscape(s), "+", "%20", -1)
}

// let's pass in URI and query as a map
// => sdb.amazonaws.com URI ? query map
// TODO: handle non URI to /
func NewSDBRequest(httpVerb string, query Strmap) *SDBRequest {
	//s := NewS3Request(httpVerb, uri)
	r := &SDBRequest{httpVerb: httpVerb, query: query}

	return r
}

func (r *SDBRequest) canonicalizedQueryString() string {
	// required AWSAccessKeyId
	// Timestamp
	// Signature. surely not here

	// a. sort
	// The parameters can come from the GET URI or from the POST body (when Content-Type is application/x-www-form-urlencoded).

	keys := SortedKeys(r.query)

	s := ""

	// b. TODO: URL encode
	for i, k := range keys {
		if 0 != i {
			s += "&"
		}
		s += percent_encode(k) + "=" + percent_encode(r.query[k])
	}

	// add SignatureVersion = "2"
	// SignatureMethod = HmacSHA256
	return s
}

func (r *SDBRequest) stringToSign() string {
	return r.httpVerb + "\n" +
		strings.ToLower("sdb.amazonaws.com") + "\n" +
		"/" + "\n" +
		r.canonicalizedQueryString()
}
