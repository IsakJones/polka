package client

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

const contentType = "application/json"

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

// RequestSnapshot sends a get request to the cache expecting all current interbank and intrabank balances.
func RequestSnapshot() ([]byte, error) {

	// Get request a snapshot of all balances
	resp, err := c.Client.Get(c.destUrl)
	if err != nil {
		return nil, err
	}

	// Read body, but keep it in json to be forwarded
	defer resp.Body.Close()
	snapshot, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

// Settle requests that the cache settle payments.
func Settle(payload *bytes.Buffer) (err error) {
	_, err = c.Client.Post(c.destUrl, c.content, payload)
	return
}
