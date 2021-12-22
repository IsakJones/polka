package client

import (
	"bytes"
	"net/http"
	"time"
)

const contentType = "transaction/json"

type client struct {
	Client  *http.Client
	destUrl string
	content string
}

var c *client

func New(destUrls []string, connTimeout, reqTimeout time.Duration) (err error) {
	transport := &http.Transport{
		MaxIdleConns:    100,
		IdleConnTimeout: connTimeout,
	}

	httpClient := &http.Client{
		Timeout:   reqTimeout,
		Transport: transport,
	}

	c = &client{
		Client:  httpClient,
		destUrl: destUrl,
		content: contentType,
	}

	return
}

func SendTransactionUpdate(payload *bytes.Buffer) (err error) {

	_, err = c.Client.Post(c.destUrl, c.content, payload)
	return
}
