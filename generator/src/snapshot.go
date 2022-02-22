package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const contentType = "application/json"

// SnapBank stores bank data relevant to a snapshot.
// SnapBank does not include the bank's id or name.
type SnapBank struct {
	Balance  int64
	Accounts map[uint32]int32
}

// Snapshot stores a synchronized snapshort of all balances.
// It stores integers and not pointers, since there's no need
// for concurrent access.
type Snapshot map[string]*SnapBank

func getSnapshot(dest string) (Snapshot, error) {
	var snap Snapshot

	// Initialize custom client
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Send get request to settler
	resp, err := c.Get(dest)
	if err != nil {
		return nil, err
	}

	// Decode body
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&snap)

	return snap, err
}

func settleBalances(dest string) error {
	var payload *bytes.Buffer

	// Initialize custom client
	c := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := c.Post(dest, contentType, payload)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		errStr := fmt.Sprintf("bad response status code: %d", resp.StatusCode)
		return errors.New(errStr)
	}

	return nil
}
