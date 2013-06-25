package s3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net/url"
	"sort"
	// "strconv"
	"strings"
	"time"
)

type S3Request struct {
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	httpVerb           string
	args               map[string]string
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
func NewS3Request(httpVerb, resource string) *S3Request {
	m := map[string]string{
		// TODO: move to test code
		"Content-MD5":  "",
		"Content-Type": "",
		// this looks right
		"Date": time.Now().Format(time.RFC1123Z),
	}

	req := S3Request{httpVerb: httpVerb, args: m, resource: resource}
	return &req
}

// Signature = Base64( HMAC-SHA1( YourSecretAccessKeyID, UTF-8-Encoding-Of( StringToSign ) ) );
// TODO: check if this is UTF8 encoding
func (req *S3Request) Signature() string {
	return SignWithKey(req.StringToSign(), req.AWSSecretAccessKey)
}

func (req *S3Request) AuthorizationString() string {
	return "AWS" + " " + req.AWSAccessKeyId + ":" + req.Signature()
}

func (req *S3Request) CanonicalizedAmzHeaders() string {
	/* oops - this is CanonicalizedResource
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
	for k, v := range req.args {
		//log.Println(strings.ToLower(k) + ":" + v)
		lower_k := strings.ToLower(k)
		if strings.HasPrefix(lower_k, "x-amz-") {
			m[lower_k] = v
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

// appends duplicate keys
func (req *S3Request) AddHeader(key, value string) {
	if prev_val, ok := req.args[key]; ok {
		// append
		value = prev_val + "," + value
	}
	req.args[key] = value
}

// rule 2 of "Constructing the CanonicalizedResource Element
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

func (req *S3Request) CanonicalizedResource() string {
	cmps := strings.Split(req.resource, "?")
	// 1. empty string
	// 2. virtual hosted bucket vs path style
	s := hostToResource(req.args["Host"]) +
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

func (req *S3Request) StringToSign() string {
	return req.httpVerb + "\n" + req.args["Content-MD5"] + "\n" + req.args["Content-Type"] + "\n" +
		req.args["Date"] + "\n" +
		req.CanonicalizedAmzHeaders() + req.CanonicalizedResource()
}

func (s3 *S3Request) AddCredentials(cred *SecurityCredentials) {
	s3.AWSAccessKeyId = cred.AWSAccessKeyId
	s3.AWSSecretAccessKey = cred.AWSSecretAccessKey
	if len(cred.token) > 0 {
		s3.args["x-amz-security-token"] = cred.token
	} else {
		log.Println("TODO: remove token here")
	}
}

// need this in (external) put code atm
func (s3 *S3Request) GetArgs() *map[string]string {
	return &s3.args
}
