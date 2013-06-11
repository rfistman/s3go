package s3

// test that we match examples found here
// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html

import (
	//	"log"
	"testing"
)

func NewR(httpVerb, date string) *S3Request {
	AWSAccessKeyId := "AKIAIOSFODNN7EXAMPLE"
	AWSSecretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

	// get request
	req := NewS3Request(httpVerb, "johnsmith", "/photos/puppy.jpg")
	req.AWSAccessKeyId = AWSAccessKeyId
	req.AWSSecretAccessKey = AWSSecretAccessKey
	req.args["Date"] = date

	return req
}

func DoTestRequest(t *testing.T, req *S3Request, e map[string]string) {
	if e["StringToSign"] != req.StringToSign() {
		t.Error(req.httpVerb + "StringToSign mismatch")
	}

	if e["Signature"] != req.Signature() {
		t.Error("signature mismatch")
	}

	if e["AuthorizationString"] != req.AuthorizationString() {
		t.Error("authorization string mismatch")
	}
}

// do I really have to do camelcase?
func Test_foo(t *testing.T) {
	// Example Object GET
	req := NewR("GET", "Tue, 27 Mar 2007 19:36:42 +0000")

	m := map[string]string{
		"StringToSign":        "GET\n\n\nTue, 27 Mar 2007 19:36:42 +0000\n/johnsmith/photos/puppy.jpg",
		"Signature":           "bWq2s1WEIj+Ydj0vQ697zp+IXMU=",
		"AuthorizationString": `AWS AKIAIOSFODNN7EXAMPLE:bWq2s1WEIj+Ydj0vQ697zp+IXMU=`,
	}

	DoTestRequest(t, req, m)

	// Example Object PUT
	// NB Content-MD5 omitted
	req = NewR("PUT", "Tue, 27 Mar 2007 21:15:45 +0000")
	req.args["Content-Type"] = "image/jpeg"

	m = map[string]string{
		"StringToSign":        "PUT\n\nimage/jpeg\nTue, 27 Mar 2007 21:15:45 +0000\n/johnsmith/photos/puppy.jpg",
		"Signature":           "MyyxeRY7whkBe+bq8fHCL/2kKUg="
		"AuthorizationString": "AWS AKIAIOSFODNN7EXAMPLE:MyyxeRY7whkBe+bq8fHCL/2kKUg=",
	}

	DoTestRequest(t, req, m)
}
