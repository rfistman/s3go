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

	s3, _ := NewS3Request("GET", "/", "") // NB: no bucket. what's this test for?
	s3.Header.Set("Host", "s3.amazonaws.com")
	s3.AddCredentials(cred)

	req, err := S3ToHttpRequest(s3, nil)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	log.Println(string(contents), err)
}
