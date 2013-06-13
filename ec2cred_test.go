package s3

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func Test_Sumpin(t *testing.T) {
	// t.Error("test sumpin")
	cred, err := GetEC2Credentials("blob_test_rw")

	if err != nil {
		t.Error(err)
		return
	}

	s3 := NewS3Request("GET", "/")
	s3.AWSAccessKeyId = cred.AWSAccessKeyId
	s3.AWSSecretAccessKey = cred.AWSSecretAccessKey
	s3.args["x-amz-security-token"] = cred.token
	s3.args["Host"] = "s3.amazonaws.com"

	// could add it to map, but that would change this
	auth := s3.AuthorizationString()

	req, err := http.NewRequest(s3.httpVerb, "http://"+s3.args["Host"]+s3.resource, nil)
	if err != nil {
		t.Error(err)
		return
	}

	for k, v := range s3.args {
		req.Header.Add(k, v)
	}
	req.Header.Add("Authorization", auth)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	log.Println(string(contents), err)
}
