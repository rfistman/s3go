package s3

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func disabledTest_Sumpin(t *testing.T) {
	// t.Error("test sumpin")
	cred, err := GetEC2Credentials("blob_test_rw")

	if err != nil {
		t.Error(err)
		return
	}

	req, _ := NewS3Request("GET", "/", "", nil) // NB: no bucket. what's this test for?
	req.Header.Set("Host", "s3.amazonaws.com")
	Authenticate(req, cred)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	log.Println(string(contents), err)
}
