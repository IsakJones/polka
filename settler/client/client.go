package client

import (
	"bytes"
	"io"
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

func New(destUrl string, reqTimeout time.Duration) (err error) {
	transport := &http.Transport{
		MaxIdleConns: 100,
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

// RequestClear sends a get request to the cache expecting all current interbank and intrabank balances.
func RequestClear() ([]byte, error) {

	resp, err := c.Client.Get(c.destUrl)
	if err != nil {
		return nil, err
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func ExecClear(payload *bytes.Buffer) (err error) {

	_, err = c.Client.Post(c.destUrl, c.content, payload)
	return
}
