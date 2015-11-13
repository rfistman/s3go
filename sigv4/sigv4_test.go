package sigv4

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
)

// http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func TestCanonicalHeaders(t *testing.T) {
	// example, removed "\n"s
	e := `content-type:application/x-www-form-urlencoded; charset=utf-8
		host:iam.amazonaws.com
		my-header1:a b c
		my-header2:"a   b   c"
		x-amz-date:20110909T233600Z`

	r, _ := http.NewRequest("GET", "http://foo.com", nil)
	r.Header.Add("Host", "iam.amazonaws.com")
	r.Header.Add("Content-type", "application/x-www-form-urlencoded; charset=utf-8")
	r.Header.Add("My-header1", "    a   b   c ")
	r.Header.Add("x-amz-date", "20110909T233600Z")
	r.Header.Add("My-Header2", "    \"a   b   c\"")

	s, _ := canonicalHeaders(r)
	if s != e {
		t.Errorf("canonical header mismatch: %v\n, %v\n", s, e)
	}
}

// generated with java example available here
// http://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-examples-using-sdks.html
func ExampleCanonicalGetRequest() {
	r, _ := http.NewRequest("GET", "http://johnsmith.s3.amazonaws.com/ExampleObject.txt", nil)
	r.Header.Add("Host", "johnsmith.s3.amazonaws.com")
	r.Header.Add("X-AMZ-Content-SHA256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	r.Header.Add("X-amz-Date", "20151110T033429Z")
	cr, _ := canonicalRequest(r)
	fmt.Printf("%v", cr)

	// Output: GET
	// /ExampleObject.txt
	//
	// host:johnsmith.s3.amazonaws.com
	// x-amz-content-sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	// x-amz-date:20151110T033429Z
	//
	// host;x-amz-content-sha256;x-amz-date
	// e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
}

// From http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func ExampleCanonicalPostRequest() {
	r, _ := http.NewRequest("POST", "http://iam.amazonaws.com", bytes.NewBufferString("Action=ListUsers&Version=2010-05-08"))
	r.Header.Add("Host", "iam.amazonaws.com")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	r.Header.Add("X-AMZ-Date", "20110909T233600Z")

	cr, _ := canonicalRequest(r)
	fmt.Printf("%v", cr)

	// Output: POST
	// /
	//
	// content-type:application/x-www-form-urlencoded; charset=utf-8
	// host:iam.amazonaws.com
	// x-amz-date:20110909T233600Z
	//
	// content-type;host;x-amz-date
	// b6359072c78d70ebee1e81adcbab4f01bf2c23245fa365ef83fe8f1f955085e2
}
