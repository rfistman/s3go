package s3

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"../util"
	"./sigv4"
)

// http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/endpoints.html

// TODO: parse ErrorResponse
// TODO: visibility timeout. for now, delete immediately?
// TODO: SetAttributes

const sqsVersion = "2012-11-05"

type SQSQueue struct {
	endpoint string
	region   string

	cred SecurityCredentials
}

// TODO: CreatQueue
type Message struct {
	Body          string
	ReceiptHandle string
	MD5OfBody     string
	MessageId     string
}

// doesn't create the queue, btw
func NewQueue(endpoint, region string) *SQSQueue {
	return &SQSQueue{endpoint: endpoint, region: region}
}

// returns nil, nil for no available message
func (sqs *SQSQueue) ReceiveMessage() (*Message, error) {
	params := url.Values{}
	params.Add("Action", "ReceiveMessage")
	// params.Add("MaxNumberOfMessages", "10")	// enable this and change to an array below
	// long polling, 20s is the maximum I think
	params.Add("WaitTimeSeconds", "20") // TODO: not here, pass it or SetAttributes

	var res struct {
		ReceiveMessageResult struct {
			Message *Message // pointer, because not present for empty queue
		}
	}

	err := sqs.doAction(&params, &res)
	if err != nil {
		return nil, err
	}
	// TODO: check MD5 of body? meh

	return res.ReceiveMessageResult.Message, nil
}

func (sqs *SQSQueue) SendMessage(message string) error {
	// TODO: more control
	// http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/AboutTimestamp.html
	// e.g. 2007-01-31T23:59:59Z -> "2006-01-02T15:04:05Z"
	//	expiryDate := time.Now().Add(5 * time.Minute)
	//	params.Add("Expires", expiryDate.Format("2006-01-02T15:04:05Z"))
	params := url.Values{}
	params.Add("Action", "SendMessage")
	params.Add("MessageBody", message)

	err := sqs.doAction(&params, nil)
	if err != nil {
		return err
	}

	return nil
}
func (sqs *SQSQueue) DeleteMessage(message *Message) error {
	params := url.Values{}
	params.Add("Action", "DeleteMessage")
	params.Add("ReceiptHandle", message.ReceiptHandle)
	return sqs.doAction(&params, nil)
}

func (sqs *SQSQueue) GetAttributes() error {
	params := url.Values{}
	params.Add("Action", "GetQueueAttributes")
	return sqs.doAction(&params, nil)
}

func (sqs *SQSQueue) doAction(params *url.Values, out interface{}) error {
	u, err := url.Parse(sqs.endpoint)
	if err != nil {
		return err
	}

	var req *http.Request

	params.Add("Version", sqsVersion)
	params.Add("QueueUrl", sqs.endpoint)

	u.RawQuery = params.Encode()

	if false {
		// GET not working. disappointing, would like signed url
		req, err = http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return err
		}
		// req.Header.Add("x-amz-content-sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
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

	if len(sqs.cred.token) > 0 {
		// needed when using ec2 instance roles
		req.Header.Add("x-amz-security-token", sqs.cred.token) // ec2 instance roles
	}

	sigv4.AuthorizeRequest(req, sqs.cred.AWSAccessKeyId, sqs.cred.AWSSecretAccessKey, sqs.region, "sqs")

	if true {
		util.LogReqAsCurl(req)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// fmt.Printf("%v", string(b))
	if resp.StatusCode != http.StatusOK {
		// TODO: parse error response?
		return errors.New(resp.Status)
	}

	if out != nil {
		err = xml.Unmarshal(b, out)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sqs *SQSQueue) AddCredentials(cred *SecurityCredentials) {
	sqs.cred = *cred // copy
}
