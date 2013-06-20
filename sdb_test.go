package s3

import (
	"log"
	"net/url"
	"strings"
	"testing"
)

var p = "https://sdb.amazonaws.com/?Action=PutAttributes&DomainName=MyDomain&ItemName=Item123&Attribute.1.Name=Color&Attribute.1.Value=Blue&Attribute.2.Name=Size&Attribute.2.Value=Med&Attribute.3.Name=Price&Attribute.3.Value=0014.99&Version=2009-04-15&Timestamp=2010-01-25T15%3A01%3A28-07%3A00&SignatureVersion=2&SignatureMethod=HmacSHA256&AWSAccessKeyId=<Your AWS Access Key ID>"

var bah = "GET\nsdb.amazonaws.com\n/\nAWSAccessKeyId=<Your AWS Access Key ID>&Action=PutAttributes&Attribute.1.Name=Color&Attribute.1.Value=Blue&Attribute.2.Name=Size&Attribute.2.Value=Med&Attribute.3.Name=Price&Attribute.3.Value=0014.99&DomainName=MyDomain&ItemName=Item123&SignatureMethod=HmacSHA256&SignatureVersion=2&Timestamp=2010-01-25T15%3A01%3A28-07%3A00&Version=2009-04-15"

func subkey(s string) string {
	return strings.Replace(s, "<Your AWS Access Key ID>", "012345", -1)
}

func TestSDBSign(t *testing.T) {
	// Example PutAttributes request

	url, err := url.Parse(subkey(p))
	if err != nil {
		t.Error(err)
	}

	//log.Println("query: ", url.RawQuery)
	// map go's richer query and map it to my own
	q := url.Query()

	m := Strmap{}

	for k, v := range q {
		//log.Println(k, v)
		if len(v) > 1 {
			log.Fatal("not expecting multi queries")
		}
		m[k] = v[0]
	}
	// log.Println(m)

	r := NewSDBRequest(m)

	// log.Println("A:", r.stringToSign(), "\nB: ", subkey(bah))

	if subkey(r.stringToSign()) != subkey(bah) {
		t.Error("string to sign mismatch")
	}
}
