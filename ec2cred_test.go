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
	s3.AddCredentials(cred)

	req, err := S3ToHttpRequest(s3)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	log.Println(string(contents), err)
}
