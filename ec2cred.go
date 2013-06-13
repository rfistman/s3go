package s3

import (
	"io/ioutil"
	"log"
	"net/http"
)

func GetEC2Credentials(role string) {
	url := "http://169.254.169.254/latest/meta-data/iam/security-credentials/" + role
	log.Println(url)
	//req, err := http.NewRequest("GET", url, nil)
	res, err := http.Get(url)
	if err != nil {
		log.Println("get error: ", err)
		return
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("read error: ", err)
		return
	}
	log.Println("CREDS: ", string(contents))
}
