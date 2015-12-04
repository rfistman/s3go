package s3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type S3Request struct {
	*http.Request

	AWSAccessKeyId     string
	AWSSecretAccessKey string
	resource           string
}

// TODO: factorise (tweet)
func B64_encode(b []byte) string {
	res := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(res, b)
	return string(res)
}

// TODO: factorise (tweet)
func SignWithKey(data, key string) string {
	// HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	hash := mac.Sum(nil)

	return B64_encode(hash)
}

// TODO: factorise (tweet)
func SortedKeys(m map[string]string) []string {
	sorted_keys := make([]string, len(m))
	i := 0
	for k, _ := range m {
		sorted_keys[i] = k
		i++
	}
	sort.Strings(sorted_keys)
	return sorted_keys
}

// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html
func NewS3Request(httpVerb, resource, bucket string) (*S3Request, error) {
	host := bucket + ".s3.amazonaws.com"
	// TODO:
	r, err := http.NewRequest(httpVerb, "https://"+host+resource, nil) // TODO: body
	if err != nil {
		return nil, err
	}

	req := S3Request{Request: r, resource: resource}
	req.Method = httpVerb
	//req.Header = make(http.Header) // TODO: REMOVE
	req.Header.Set("Host", host)
	req.Header.Set("Date", time.Now().Format(time.RFC1123Z))
	return &req, nil
}

// Signature = Base64( HMAC-SHA1( YourSecretAccessKeyID, UTF-8-Encoding-Of( stringToSign ) ) );
// TODO: check if this is UTF8 encoding
func (req *S3Request) signature() string {
	return SignWithKey(req.stringToSign(), req.AWSSecretAccessKey)
}

func (req *S3Request) authorizationString() string {
	return "AWS" + " " + req.AWSAccessKeyId + ":" + req.signature()
}

func (req *S3Request) canonicalizedAmzHeaders() string {
	/* oops - this is canonicalizedResource
	// 1. start with an empty string
	s := ""

	// 2. virtual hosting style buckets??? (unimplemented)
	// nothing for path style requests

	// 3. append the path of the undecoded http request-URI,
	// up to but not including the query string
	return s
	*/

	// 1. convert each http headername to lower case
	m := map[string]string{}
	for k, vs := range req.Header {
		//log.Println(strings.ToLower(k) + ":" + v)
		lower_k := strings.ToLower(k)
		if strings.HasPrefix(lower_k, "x-amz-") {
			m[lower_k] = strings.Join(vs, ",")
		}
	}

	// 2. sort collection of headers lexographically by header name
	sorted_keys := SortedKeys(m)

	// 3. combine same name header fields (already done with AddHeader)

	// TODO: 4. unfold multiline headers
	// 5. remove white space around colon (or don't add it)

	s := ""
	for _, k := range sorted_keys {
		s += k + ":" + m[k] + "\n"
	}

	return s
}

// rule 2 of "Constructing the canonicalizedResource Element
func hostToResource(host string) string {
	// path style, I guess
	if host == "s3.amazonaws.com" {
		return ""
	}

	// virtual hosted-style request
	// remove s3 aws if present
	suffix := ".s3.amazonaws.com"
	if strings.HasSuffix(host, suffix) {
		host = host[:len(host)-len(suffix)]
	} else {
		// remove :port
		host = strings.Split(host, ":")[0]
	}
	// return custom host thing? I don't see where this is documented
	// for now matching example
	return "/" + host
}

var sorted_included_sub_resources = []string{"acl", "lifecycle", "location", "logging",
	"notification", "partNumber", "policy", "requestPayment", "torrent", "uploadId",
	"uploads", "versionId", "versioning", "versions", "website"}

func getIncludedQuery(query string) string {
	// TODO: handle error. pass it back?
	m, err := url.ParseQuery(query)
	if err != nil {
		log.Println("ParseQuery: ", err)
		return ""
	}

	// included_values := url.Values{}

	s := ""
	for _, k := range sorted_included_sub_resources {
		if v_arr, ok := m[k]; ok {
			if len(s) > 0 {
				s += "&"
			}
			s += k
			for i, v := range v_arr {
				// included_values.Add(k, v)
				if i != 0 {
					s += k
				}
				// ParseQuery gives me one zero length string for ?acl
				if len(v) > 0 {
					s += "=" + v
				}
			}
		}
	}
	// encodes no value as "k="
	//return included_values.Encode()
	return s
}

func (req *S3Request) canonicalizedResource() string {
	cmps := strings.Split(req.resource, "?")
	// 1. empty string
	// 2. virtual hosted bucket vs path style
	s := hostToResource(req.Header.Get("Host")) +
		// 3. path part up to but not including query string
		cmps[0]

	// 4. included sub-resources
	// TODO: response header overrides
	if len(cmps) > 1 {
		query := cmps[1]
		included_query := getIncludedQuery(query)
		if len(included_query) > 0 {
			s += "?" + included_query
		}
	}
	return s
}

func (req *S3Request) stringToSign() string {
	h := req.Header
	return req.Method + "\n" + h.Get("Content-MD5") + "\n" + h.Get("Content-Type") + "\n" +
		h.Get("Date") + "\n" +
		req.canonicalizedAmzHeaders() + req.canonicalizedResource()
}

func (s3 *S3Request) Authenticate(cred *SecurityCredentials) {
	s3.Header.Set("x-amz-security-token", cred.token)

	// could add it to map, but that would change this
	s3.Header.Add("Authorization", s3.authorizationString())
}

func (s3 *S3Request) AddCredentials(cred *SecurityCredentials) {
	s3.AWSAccessKeyId = cred.AWSAccessKeyId
	s3.AWSSecretAccessKey = cred.AWSSecretAccessKey
	if len(cred.token) > 0 {
		s3.Header.Set("x-amz-security-token", cred.token)
	} else {
		// log.Println("TODO: remove token here")
	}
}
