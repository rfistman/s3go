package s3

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

// http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/endpoints.html

const sqsVersion = "2012-11-05"

type SQSQueue struct {
	endpoint string
}

// TODO: CreatQueue

func NewQueue(endpoint string) *SQSQueue {
	return &SQSQueue{endpoint: endpoint}
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
	// TODO: more control
	// http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/AboutTimestamp.html
	// e.g. 2007-01-31T23:59:59Z -> "2006-01-02T15:04:05Z"
	expiryDate := time.Now().Add(5 * time.Minute)
	params.Add("Expires", expiryDate.Format("2006-01-02T15:04:05Z"))
	u.RawQuery = params.Encode()
	log.Printf("FF %v", u.String())

	r, _ := http.NewRequest("GET", "http://bler.com/foo?gah=1234&boot=a ?", nil)
	r.Header.Add("Host", "iam.amazonaws.com") // REQUIRED!
	r.Header.Add("x-amz-date", "asdfasdf")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	return nil
}
