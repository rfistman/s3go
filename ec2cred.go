package s3

import (
	"encoding/json"
	"io/ioutil"
	//"log"
	"net/http"
)

var credsUrl = "http://169.254.169.254/latest/meta-data/iam/security-credentials/"

type SecurityCredentials struct {
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	token              string
}

func GetEC2Role() (string, error) {
	res, err := http.Get(credsUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	// looks like there's only ever one role
	return string(contents), nil
	//return strings.Split(string(contents), "\n"), nil
}

func GetEC2Credentials(role string) (*SecurityCredentials, error) {
	url := credsUrl + role
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
