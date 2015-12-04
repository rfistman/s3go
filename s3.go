package s3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// TODO: factorise (tweet)
func B64_encode(b []byte) string {
	res := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(res, b)
	return string(res)
}

// TODO: factorise (tweet)
func signWithKey(data, key string) string {
	// HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(data))
	hash := mac.Sum(nil)

	return B64_encode(hash)
}

// TODO: factorise (tweet)
func sortedKeys(m map[string]string) []string {
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
func NewS3Request(httpVerb, resource, bucket string, body io.Reader) (*http.Request, error) {
	host := bucket + ".s3.amazonaws.com"
	req, err := http.NewRequest(httpVerb, "https://"+host+resource, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", host)
	req.Header.Set("Date", time.Now().Format(time.RFC1123Z))
	return req, nil
}

// Signature = Base64( HMAC-SHA1( YourSecretAccessKeyID, UTF-8-Encoding-Of( stringToSign ) ) );
// TODO: check if this is UTF8 encoding
func signature(req *http.Request, awsSecretAccessKey string) string {
	return signWithKey(stringToSign(req), awsSecretAccessKey)
}

func authorizationString(req *http.Request, cred *SecurityCredentials) string {
	return "AWS" + " " + cred.AWSAccessKeyId + ":" + signature(req, cred.AWSSecretAccessKey)
}

func canonicalizedAmzHeaders(req *http.Request) string {
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
	sorted_keys := sortedKeys(m)

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

func canonicalizedResource(req *http.Request) string {
	// 1. empty string
	// 2. virtual hosted bucket vs path style
	s := hostToResource(req.Header.Get("Host")) +
		// 3. path part up to but not including query string
		req.URL.Path

	// 4. included sub-resources
	// TODO: response header overrides
	if len(req.URL.RawQuery) > 1 {
		included_query := getIncludedQuery(req.URL.RawQuery)
		if len(included_query) > 0 {
			s += "?" + included_query
		}
	}
	return s
}

func stringToSign(req *http.Request) string {
	h := req.Header
	return req.Method + "\n" + h.Get("Content-MD5") + "\n" + h.Get("Content-Type") + "\n" +
		h.Get("Date") + "\n" +
		canonicalizedAmzHeaders(req) + canonicalizedResource(req)
}

func Authenticate(req *http.Request, cred *SecurityCredentials) {
	req.Header.Set("x-amz-security-token", cred.token)

	// could add it to map, but that would change this
	req.Header.Add("Authorization", authorizationString(req, cred))
}
