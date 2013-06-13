package s3

import (
	"encoding/json"
	"io/ioutil"
	//"log"
	"net/http"
)

type SecurityCredentials struct {
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	token              string
}

func GetEC2Credentials(role string) (*SecurityCredentials, error) {
	url := "http://169.254.169.254/latest/meta-data/iam/security-credentials/" + role
	//url := "http://localhost:1234/latest/meta-data/iam/security-credentials/" + role
	//req, err := http.NewRequest("GET", url, nil)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	//log.Println("CREDS: ", string(contents))
	var m map[string]string
	err = json.Unmarshal(contents, &m)
	//log.Println(err, m)

	return &SecurityCredentials{m["AccessKeyId"], m["SecretAccessKey"], m["Token"]}, nil
}
