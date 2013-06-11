package s3

import (
	//	"strconv"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"time"
)

type S3Request struct {
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	httpVerb           string
	args               map[string]string
	bucket             string
	resource           string
}

// TODO: factorise (tweet)
func b64_encode(b []byte) string {
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

	return b64_encode(hash)
}

// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html
func NewS3Request(httpVerb, bucket, resource string) *S3Request {
	m := map[string]string{
		"Host":         bucket + ".s3.amazonaws.com",
		"Content-MD5":  "",
		"Content-Type": "",
		// this looks right
		"Date": time.Now().Format(time.RFC1123Z),
	}

	req := S3Request{httpVerb: httpVerb, args: m, bucket: bucket, resource: resource}
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
	// unimplemented
	return ""
}

func (req *S3Request) CanonicalizedResource() string {
	return "/" + req.bucket + req.resource
}

func (req *S3Request) StringToSign() string {
	return req.httpVerb + "\n" + req.args["Content-MD5"] + "\n" + req.args["Content-Type"] + "\n" +
		req.args["Date"] + "\n" +
		req.CanonicalizedAmzHeaders() + req.CanonicalizedResource()
}
