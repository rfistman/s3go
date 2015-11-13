package sigv4

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"../../util"
)

func hmacSha256(data string, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data)) // hash.Hash writer never returns an error
	return mac.Sum(nil)
}

// supposed to be lower case. if hex stops doing this add ToLower
func hexSha256(data []byte) string {
	s := sha256.New()
	s.Write(data)
	return hex.EncodeToString(s.Sum(nil))
}

// http://docs.aws.amazon.com/general/latest/gr/sigv4-calculate-signature.html
// http://docs.aws.amazon.com/general/latest/gr/signature-v4-examples.html
func signingKey(key, dateStamp, regionName, serviceName string) []byte {
	kSecret := []byte("AWS4" + key)
	kDate := hmacSha256(dateStamp, kSecret)
	kRegion := hmacSha256(regionName, kDate)
	kService := hmacSha256(serviceName, kRegion)
	kSigning := hmacSha256("aws4_request", kService)

	return kSigning
}

// http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func canonicalRequest(req *http.Request) (s string, err error) {
	// 1. request method
	s = req.Method + "\n"

	// 2. canonical URI
	canonicalURI := req.URL.Path
	if len(canonicalURI) == 0 {
		canonicalURI = "/"
	}
	s += canonicalURI + "\n"

	// 3. canonical query string
	cq, err := canonicalQueryString(req.URL.RawQuery)
	if err != nil {
		return
	}
	s += cq + "\n"

	// 4. canonical headers
	ch, signedHeaders := canonicalHeaders(req)
	s += ch + "\n"

	// 5. add signed headers. seems like canonicalHeaders would know about this
	s += strings.Join(signedHeaders, ";") + "\n"

	// BUG: don't read it all in
	var body []byte // empty

	if req.Body != nil {
		// BLEAH. copy body and recreate it.
		// TODO: push hash back to user
		body, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return
		}
		req.Body.Close() // so bad
		req.Body = util.NopCloser{bytes.NewBuffer(body)}
	}

	// 6. hashed payload. I'm doing empty for now
	s += hexSha256(body) // NB: no trailing newline!
	return
}

// your raw query might already be in this form via url.Values.Encode, but best to be sure
func canonicalQueryString(rawQuery string) (string, error) {
	m, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", err
	}
	// This percent encodes and sorts by key! which appears to cover steps a-e!
	// http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
	return m.Encode(), nil
}

// returns the canonical headers and signed headers (sorted, lowercased)
func canonicalHeaders(req *http.Request) (ch string, signedHeaders []string) {
	m := make(map[string]string)

	// There's fkn quoting
	// spaceSucker, _ := regexp.Compile(" +")

	for k, v := range req.Header {
		// TODO: map consecutive spaces in v to one
		// aka http://tools.ietf.org/html/rfc2616#page-32
		// TODO: figure out what multiple values mean
		ck := strings.ToLower(k)
		signedHeaders = append(signedHeaders, ck)
		m[ck] = strings.TrimSpace(v[0])
		// m[ck] = spaceSucker.ReplaceAllString(strings.TrimSpace(v[0]), " ")
	}
	sort.Strings(signedHeaders)
	for _, k := range signedHeaders {
		ch += k + ":" + m[k] + "\n"

	}
	return
}

// http://docs.aws.amazon.com/general/latest/gr/sigv4-create-string-to-sign.html
func stringToSign(req *http.Request, dateISO8601, region, service string) (s string, err error) {
	s = "AWS4-HMAC-SHA256\n" // 1. the algorithm. all sha256 for now
	s += dateISO8601 + "\n"  // 2. date

	var YYYYMMDD string
	// pull YYYYMMDD out of date. hacky?
	if len(dateISO8601) >= 8 {
		YYYYMMDD = dateISO8601[:8]
	} else {
		err = errors.New("invalid ISO8601 date")
		return
	}
	s += fmt.Sprintf("%v/%v/%v/aws4_request", YYYYMMDD, region, service) + "\n" // 3. credential scope

	var cr string
	cr, err = canonicalRequest(req)
	if err != nil {
		return
	}
	s += hexSha256([]byte(cr)) // 4. hash of canonical request. WITHOUT newline
	return
}
