package s3

import (
	// "log"
	"net/url"
	"strings"
	"time"
)

// TODO: POST with Content-Type is application/x-www-form-urlencoded?
// http://docs.aws.amazon.com/AmazonSimpleDB/latest/DeveloperGuide/HMACAuth.html

type AWSRequest interface {
	StringToSign() string
	AddCredentials(cred *SecurityCredentials)
}

type Strmap map[string]string

type SDBRequest struct {
	httpVerb string
	// uri      string	// hardcoding as "/"
	query              Strmap
	AWSSecretAccessKey string // don't actually want this exported
	host               string
}

// TODO: factorise (tweet)
func percent_encode(s string) string {
	// go's "+" encoding for space doesn't seem to please
	//return url.QueryEscape(s)
	return strings.Replace(url.QueryEscape(s), "+", "%20", -1)
}

// let's pass in URI and query as a map
// => sdb.amazonaws.com URI ? query map
// TODO: handle non URI to /
func NewSDBRequest(query Strmap) *SDBRequest {
	// wrong timestamp? should be other format
	// http://docs.aws.amazon.com/AmazonSimpleDB/latest/DeveloperGuide/HMACAuth.html#AboutTimestamp
	m := Strmap{
		// e.g 2010-01-31T23:59:59Z GMT/CUT recommended but not required.
		"Timestamp":        time.Now().UTC().Format(time.RFC3339),
		"Version":          "2009-04-15",
		"SignatureVersion": "2",
		// pretty sure I'm using HmacSHA1
		"SignatureMethod": "HmacSHA1",
	}

	// let caller overwrite defaults
	for k, v := range query {
		m[k] = v
	}

	r := &SDBRequest{httpVerb: "GET", query: m, host: "sdb.amazonaws.com"}

	return r
}

func (r *SDBRequest) AddCredentials(cred *SecurityCredentials) {
	// in S3 it's in the http header, here it's in the query
	r.query["AWSAccessKeyId"] = cred.AWSAccessKeyId
	r.AWSSecretAccessKey = cred.AWSSecretAccessKey

	if len(cred.token) > 0 {
		r.query["SecurityToken"] = cred.token
	}
}

func (r *SDBRequest) canonicalizedQueryString() string {
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

	return s
}

func (r *SDBRequest) stringToSign() string {
	return r.httpVerb + "\n" +
		strings.ToLower(r.host) + "\n" +
		"/" + "\n" +
		r.canonicalizedQueryString()
}

func (r *SDBRequest) signature() string {
	return percent_encode(SignWithKey(r.stringToSign(), r.AWSSecretAccessKey))
}
