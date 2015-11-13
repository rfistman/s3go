package s3

import (
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

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Host", u.Host) // REQUIRED!

	sigv4.AuthorizeRequest(req, sqs.accessKeyId, sqs.secretAccessKey, sqs.region, "sqs")

	util.LogReqAsCurl(req)
	if false {
		c := http.Client{}
		resp, _ := c.Do(req)
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("FF %v", string(b))
	}
	return nil
}

func (sqs *SQSQueue) AddCredentials(cred *SecurityCredentials) {
	sqs.accessKeyId = cred.AWSAccessKeyId
	sqs.secretAccessKey = cred.AWSSecretAccessKey
}
