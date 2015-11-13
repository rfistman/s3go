package s3

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"../s3"
)

// TODO: move into s3

var useEnvCreds = flag.Bool("envCred", false, "use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY environment variables")

// TODO: reconcile with getCred in awsbn/sdb_play.go
func GetCred() (*s3.SecurityCredentials, error) {
	if *useEnvCreds {
		creds := &s3.SecurityCredentials{AWSAccessKeyId: os.Getenv("AWS_ACCESS_KEY_ID"), AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY")}
		return creds, nil
	}

	role, err := s3.GetEC2Role()
	if err != nil {
		return nil, err
	}
	return s3.GetEC2Credentials(role)
}

func S3Req(httpVerb, resourceName, bucket string) (*s3.S3Request, error) {
	r := s3.NewS3Request(httpVerb, resourceName)
	cred, err := GetCred()
	if err != nil {
		return nil, err
	}
	r.AddCredentials(cred)
	(*r.GetArgs())["Host"] = bucket + ".s3.amazonaws.com"
	return r, nil
}

// maybe pass fn to make request.
func DoRequest(req *http.Request) (body []byte, err error) {
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	// Probably not the way to handle both 200 and 204 (no content)?
	if 2 == response.StatusCode/100 {
		body = contents
		return // both body, and nil err
	}

	// TODO: interpret amazon xml error documents here
	// or not - they look different at http level
	err = errors.New(string(contents))
	return
}

// TODO: refactor into s3 code from awsbn
//  openssl dgst -md5 -binary | openssl enc -base64
func B64MD5(b io.Reader) (string, int64) {
	h := md5.New()
	written, err := io.Copy(h, b)

	if err != nil {
		log.Fatal(err)
	}
	hash := h.Sum(nil)
	return s3.B64_encode(hash), written
}

func S3UrlS3GetReq(s3SchemeUrl string) (*S3Req, error) {
	u, err := url.Parse(s3SchemeUrl)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "s3" {
		return nil, fmt.Errorf("non-s3 scheme: %v", u.Scheme)
	}
	s3r, err := S3Req("GET", u.Path, u.Host)
	if err != nil {
		return nil, err
	}
	return s3r, nil
}

func S3UrlGetRequest(s3SchemeUrl string) (*http.Request, error) {
	s3r, err := S3UrlS3GetReq(s3SchemeUrl)
	if err != nil {
		return nil, err
	}

	return s3.S3ToHttpRequest(s3r, nil)
}

func UnmarshalFromS3(s3url string, out interface{}) error {
	req, err := S3UrlGetRequest(s3url)
	if err != nil {
		return err
	}

	b, err := DoRequest(req)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, out)
	return err
}
