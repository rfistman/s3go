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
	cr, _, _ := canonicalRequest(r)
	fmt.Printf("%v\n", cr)

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

func samplePostRequest() *http.Request {
	r, _ := http.NewRequest("POST", "https://iam.amazonaws.com", bytes.NewBufferString("Action=ListUsers&Version=2010-05-08"))
	r.Header.Add("Host", "iam.amazonaws.com")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	r.Header.Add("X-AMZ-Date", "20110909T233600Z")

	return r
}

// From http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func ExampleCanonicalPostRequest() {
	cr, _, _ := canonicalRequest(samplePostRequest())
	fmt.Printf("%v\n", cr)

	// throw step 8 in as well
	fmt.Printf("Hashed canonical request: %v\n", hexSha256([]byte(cr)))

	// Output: POST
	// /
	//
	// content-type:application/x-www-form-urlencoded; charset=utf-8
	// host:iam.amazonaws.com
	// x-amz-date:20110909T233600Z
	//
	// content-type;host;x-amz-date
	// b6359072c78d70ebee1e81adcbab4f01bf2c23245fa365ef83fe8f1f955085e2
	// Hashed canonical request: 3511de7e95d28ecd39e9513b642aee07e54f4941150d8df8bf94b328ef7e55e2
}

// http://docs.aws.amazon.com/general/latest/gr/sigv4-create-string-to-sign.html
func ExampleSampleStringToSign() {
	s, _, _, _, _ := stringToSign(samplePostRequest(), "20110909T233600Z", "us-east-1", "iam")
	fmt.Printf("%v", s)

	// Output: AWS4-HMAC-SHA256
	// 20110909T233600Z
	// 20110909/us-east-1/iam/aws4_request
	// 3511de7e95d28ecd39e9513b642aee07e54f4941150d8df8bf94b328ef7e55e2
}

const sampleSecretAccessKey string = "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY"

func ExampleSigningKey() {
	sk := signingKey(sampleSecretAccessKey, "20110909", "us-east-1", "iam")
	fmt.Printf("signing key: %v\n", sk)

	sig, _, _, _ := signature(samplePostRequest(), sampleSecretAccessKey, "20110909T233600Z", "us-east-1", "iam")
	fmt.Printf("signature: %v\n", sig)

	// Output: signing key: [152 241 216 137 254 196 244 66 26 220 82 43 171 12 225 248 46 105 41 194 98 237 21 229 169 76 144 239 209 227 176 231]
	// signature: ced6826de92d2bdeed8f846f0bf508e8559e98e4b0199114b84c54174deb456c
}

// http://docs.aws.amazon.com/general/latest/gr/sigv4-add-signature-to-request.html
func ExampleAuthorizationHeader() {
	auth, _ := authorizationString(samplePostRequest(), "AKIDEXAMPLE", sampleSecretAccessKey, "20110909T233600Z", "us-east-1", "iam")
	fmt.Printf("Authorization: %v", auth)

	// Output: Authorization: AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20110909/us-east-1/iam/aws4_request, SignedHeaders=content-type;host;x-amz-date, Signature=ced6826de92d2bdeed8f846f0bf508e8559e98e4b0199114b84c54174deb456c

}
