package s3

// test that we match examples found here
// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html

import (
	"log"
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

	AuthorizationString := "AWS AKIAIOSFODNN7EXAMPLE:" + e["Signature"]

	if AuthorizationString != req.AuthorizationString() {
		t.Error("authorization string mismatch")
	}
}

// do I really have to do camelcase?
func Test_ObjectGET(t *testing.T) {
	// Example Object GET
	req := NewR("GET", "Tue, 27 Mar 2007 19:36:42 +0000")

	m := map[string]string{
		"StringToSign": "GET\n\n\nTue, 27 Mar 2007 19:36:42 +0000\n/johnsmith/photos/puppy.jpg",
		"Signature":    "bWq2s1WEIj+Ydj0vQ697zp+IXMU=",
	}

	DoTestRequest(t, req, m)

}

func Test_ObjectPUT(t *testing.T) {
	// Example Object PUT
	// NB Content-MD5 omitted
	req := NewR("PUT", "Tue, 27 Mar 2007 21:15:45 +0000")
	req.args["Content-Type"] = "image/jpeg"

	m := map[string]string{
		"StringToSign": "PUT\n\nimage/jpeg\nTue, 27 Mar 2007 21:15:45 +0000\n/johnsmith/photos/puppy.jpg",
		"Signature":    "MyyxeRY7whkBe+bq8fHCL/2kKUg=",
	}

	DoTestRequest(t, req, m)
}

func Test_List(t *testing.T) {
	req := NewR("GET", "Tue, 27 Mar 2007 19:42:41 +0000")
	req.resource = "/?prefix=photos&max-keys=50&marker=puppy"

	// Example list
	m := map[string]string{
		"StringToSign": "GET\n\n\nTue, 27 Mar 2007 19:42:41 +0000\n/johnsmith/",
		"Signature":    "htDYFYduRNen8P9ZfE/s9SuKy0U=",
	}
	log.Println(req.StringToSign())

	DoTestRequest(t, req, m)
}

func Test_Fetch(t *testing.T) {
	req := NewR("GET", "Tue, 27 Mar 2007 19:44:46 +0000")
	req.resource = "/?acl"

	m := map[string]string{
		"StringToSign": "GET\n\n\nTue, 27 Mar 2007 19:44:46 +0000\n/johnsmith/?acl",
		"Signature":    "c2WLPFtWHVgbEmeEG93a4cG37dM=",
	}

	DoTestRequest(t, req, m)
}

func Test_Delete(t *testing.T) {
	// NB example date is wrong in this example. should be 26s not 27s
	req := NewR("DELETE", "Tue, 27 Mar 2007 21:20:26 +0000")

	m := map[string]string{
		"StringToSign": "DELETE\n\n\nTue, 27 Mar 2007 21:20:26 +0000\n/johnsmith/photos/puppy.jpg",
		"Signature":    "lx3byBScXR6KzyMaifNkardMwNk=",
	}

	DoTestRequest(t, req, m)
}
