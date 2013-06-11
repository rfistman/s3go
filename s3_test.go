package s3

// test that we match examples found here
// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html

import (
	"log"
	"testing"
)

// do I really have to do camelcase?
func Test_foo(t *testing.T) {
	AWSAccessKeyId := "AKIAIOSFODNN7EXAMPLE"
	AWSSecretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

	req := NewS3Request("GET", "johnsmith.s3.amazonaws.com", "/johnsmith/photos/puppy.jpg")
	req.AWSAccessKeyId = AWSAccessKeyId
	req.AWSSecretAccessKey = AWSSecretAccessKey
	req.args["Date"] = "Tue, 27 Mar 2007 19:36:42 +0000"

	expectedAuthorizationString := `Authorization: AWS AKIAIOSFODNN7EXAMPLE: bWq2s1WEIj+Ydj0vQ697zp+IXMU=`
	if expectedAuthorizationString != req.AuthorizationString() {
		t.Error("authorization string mismatch")
	}
	expectedStringToSign := `GET\n
\n
\n
Tue, 27 Mar 2007 19:36:42 +0000\n
/johnsmith/photos/puppy.jpg`
	log.Println(req.StringToSign())
	if expectedStringToSign != req.StringToSign() {
		t.Error("string to sign mismatch")
	}

}
