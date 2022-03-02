package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sekerez/polka/utils"
)

const contentType = "application/json"

func getSnapshot(dest string) (*utils.Snapshot, error) {
	var snap utils.Snapshot

	// Initialize custom client
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Send get request to settler
	resp, err := c.Get(dest)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		errStr := fmt.Sprintf("bad response status code: %d", resp.StatusCode)
		return nil, errors.New(errStr)
	}

	// byteBody, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// fmt.Printf("%s\n", byteBody)

	// return nil, nil

	// Decode body
	err = json.NewDecoder(resp.Body).Decode(&snap)
	if err != nil {
		return nil, err
	}

	if snap.Banks == nil {
		return nil, errors.New("Got nil snapshot")
	}
	snap.Print()

	return &snap, err
}

func settleBalances(dest string) error {
	var payload bytes.Buffer

	// Initialize custom client
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := c.Post(dest, contentType, &payload)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		errStr := fmt.Sprintf("bad response status code: %d", resp.StatusCode)
		return errors.New(errStr)
	}

	return nil
}
