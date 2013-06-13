package s3

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func GetEC2Credentials(role string) {
	url := "http://169.254.169.254/latest/meta-data/iam/security-credentials/" + role
	//url := "http://localhost:1234/latest/meta-data/iam/security-credentials/" + role
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
	var objmap map[string]string
	err = json.Unmarshal(contents, &objmap)
	log.Println(err, objmap)

}
