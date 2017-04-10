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
)

// should probably default to true
// needed to make temporary credentials work (like you find on lambda)
var useEnvCreds = flag.Bool("envCred", false, "use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_SESSION_TOKEN environment variables")

// TODO: reconcile with getCred in awsbn/sdb_play.go
func GetCred() (*SecurityCredentials, error) {
	if *useEnvCreds {
		// token makes lambda work
		creds := &SecurityCredentials{AWSAccessKeyId: os.Getenv("AWS_ACCESS_KEY_ID"), AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"), token: os.Getenv("AWS_SESSION_TOKEN")}
		return creds, nil
	}

	role, err := GetEC2Role()
	if err != nil {
		return nil, err
	}
	return GetEC2Credentials(role)
}

// maybe pass fn to make request.
func DoRequest(req *http.Request) (body []byte, err error) {
	// CRAPPY
	cred, err := GetCred()
	if err != nil {
		return nil, err
	}
	Authenticate(req, cred)

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

//  openssl dgst -md5 -binary | openssl enc -base64
func B64MD5(b io.Reader) (string, int64) {
	h := md5.New()
	written, err := io.Copy(h, b)

	if err != nil {
		log.Fatal(err)
	}
	hash := h.Sum(nil)
	return B64_encode(hash), written
}

func S3UrlGetReq(s3SchemeUrl string) (*http.Request, error) {
	u, err := url.Parse(s3SchemeUrl)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "s3" {
		return nil, fmt.Errorf("non-s3 scheme: %v", u.Scheme)
	}
	req, err := NewS3Request("GET", u.Path, u.Host, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func UnmarshalFromS3(s3SchemeUrl string, out interface{}) error {
	req, err := S3UrlGetReq(s3SchemeUrl)
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
