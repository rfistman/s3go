package s3

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"../util"
	"./sigv4"
)

// http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/endpoints.html

const sqsVersion = "2012-11-05"

type SQSQueue struct {
	endpoint string
	region   string

	accessKeyId     string
	secretAccessKey string
}

// TODO: CreatQueue

func NewQueue(endpoint, region string) *SQSQueue {
	return &SQSQueue{endpoint: endpoint, region: region}
}

func (sqs *SQSQueue) SendMessage(message string) error {
	//var u url.URL
	u, err := url.Parse(sqs.endpoint)
	if err != nil {
		return err
	}

	var req *http.Request

	params := url.Values{}
	params.Add("Action", "SendMessage")
	params.Add("MessageBody", message)
	params.Add("Version", sqsVersion)
	params.Add("QueueUrl", sqs.endpoint)

	// TODO: more control
	// http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/AboutTimestamp.html
	// e.g. 2007-01-31T23:59:59Z -> "2006-01-02T15:04:05Z"
	//	expiryDate := time.Now().Add(5 * time.Minute)
	//	params.Add("Expires", expiryDate.Format("2006-01-02T15:04:05Z"))
	u.RawQuery = params.Encode()

	if false {
		// GET not working. disappointing, would like signed url
		req, err = http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return err
		}
	} else {
		// POST works!
		u.Path = "" // silly hack to get scheme://host
		u.RawQuery = ""
		req, err = http.NewRequest("POST", u.String(), bytes.NewBufferString(params.Encode()))
		if err != nil {
			return err
		}
		// without this we get the mysterious
		// Unable to determine service/operation name to be authorized
		// NB: doesn't need to be under the seal, but it's post only, so...
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	req.Header.Add("Host", u.Host) // REQUIRED!

	sigv4.AuthorizeRequest(req, sqs.accessKeyId, sqs.secretAccessKey, sqs.region, "sqs")

	if true {
		log.Printf("GUH: %+v", req)
		c := http.Client{}
		resp, _ := c.Do(req)
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("FF %v", string(b))
	} else {
		util.LogReqAsCurl(req)
	}
	return nil
}

func (sqs *SQSQueue) AddCredentials(cred *SecurityCredentials) {
	sqs.accessKeyId = cred.AWSAccessKeyId
	sqs.secretAccessKey = cred.AWSSecretAccessKey
}
