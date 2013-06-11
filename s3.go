package s3

import (
	//	"strconv"
	"time"
)

type S3Request struct {
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	httpVerb           string
	args               map[string]string
}

// http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html
func NewS3Request(httpVerb, host, name string) *S3Request {
	m := map[string]string{
		"Host":         host,
		"Content-MD5":  "",
		"Content-Type": "",
		// this looks right
		"Date": time.Now().Format(time.RFC1123Z),
	}

	req := S3Request{httpVerb: httpVerb, args: m}
	return &req
}

func (req *S3Request) Signature() string {
	return ""
}

func (req *S3Request) AuthorizationString() string {
	return "AWS" + " " + req.AWSAccessKeyId + ":" + req.Signature()
}

func (req *S3Request) StringToSign() string {
	return req.httpVerb + "\n" + req.args["Content-MD5"] + "\n" + req.args["Content-Type"] + "\n" +
		req.args["Date"]
}
