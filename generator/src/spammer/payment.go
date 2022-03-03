package spammer

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/sekerez/polka/utils"
)

const (
	lo = 0
	hi = 100000

	timeout     = 3 * time.Second
	contentType = "application/json"
)

var (
	badResponses *uint32
	c            *http.Client

	// list of the 10 largest US banks
	banks = []string{
		"JP Morgan Chase",
		"Bank of America",
		"Wells Fargo",
		"Citigroup",
		"U.S. Bancorp",
		"Truist Financial",
		"PNC Financial Services Group",
		"TD Group US",
		"Bank of New York Mellon",
		"Capital One Financial",
	}
)

// PaymentSpammer sends transactionNumber POST requests concurrently,
// to the specified dest. The goal is to send requests as
// close to simultaneous as possible. Returns number of bad requests.
func PaymentSpammer(dest string, maxGoroutines, transactionNumber uint, measure bool) uint32 {

	badResponses = new(uint32)

	workChan := make(chan interface{})
	doneChanNew := make(chan interface{})

	// Reset randomness
	rand.Seed(int64(time.Now().Second()))

	// Initialize custom client
	c = &http.Client{
		Timeout: 5 * time.Second,
	}

	// Initialize limited number of workers
	log.Printf("initializing workers")
	for i := uint(0); i < maxGoroutines; i++ {
		go Worker(lo, hi, dest, workChan, doneChanNew, measure)
	}

	// Assign work to workers
	log.Printf("filling work channel")
	for i := uint(0); i < transactionNumber; i++ {
		workChan <- true
	}

	// End workers
	log.Printf("killing workers")
	for i := uint(0); i < maxGoroutines; i++ {
		doneChanNew <- true
	}
	log.Printf("Finished.")

	return *badResponses
}

func Worker(lo, hi int, dest string, work <-chan interface{}, done <-chan interface{}, measure bool) {
	for {
		select {
		case <-done:
			return
		case <-work:
			payload := generateTransaction(lo, hi)

			if !measure {
				sendTransaction(dest, payload)
				continue
			}

			// Post request
			start := time.Now()
			sendTransaction(dest, payload)
			elapsedNum := int64(time.Since(start) / time.Nanosecond)
			elapsedBytes := []byte(strconv.FormatInt(elapsedNum, 10) + "\n")

			// Append to file
			f, err := os.OpenFile("measurements.txt", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("Error appending measurements to file: %s", err.Error())
			}
			if _, err := f.Write(elapsedBytes); err != nil {
				log.Fatalf("Error writing to measurements file: %s", err.Error())
			}
			if err := f.Close(); err != nil {
				log.Fatalf("Error writing to measurements file: %s", err.Error())
			}

			// log.Printf("Sending the transation took %s", end.Sub(start))
		}
	}
}

// sendTransaction sends a POST request with a json-encoded
// payment to the load balancer.
func sendTransaction(dest string, payload *bytes.Buffer) {
	resp, err := c.Post(dest, contentType, payload)
	if err != nil {
		log.Printf("Error sending request: %s", err.Error())
		return
	}
	defer resp.Body.Close()
	// In case of failure, print
	if resp.StatusCode > 299 {
		atomic.AddUint32(badResponses, 1)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Can't read body: %s", err.Error())
		}
		log.Printf("Received response with bad status code: %s", string(bodyBytes))
	}
}

// Returns a buffer fit for the POST request encoding
// a transaction struct.
func generateTransaction(lo, hi int) *bytes.Buffer {

	// calculate basic transaction attributes
	sum := rand.Intn(hi-lo) + lo
	time := time.Now()
	sendAcc := rand.Intn(100)
	receiverAcc := rand.Intn(100)
	senderIndex := rand.Intn(len(banks))
	receiverIndex := rand.Intn(len(banks))

	// Make sure that the sending bank and the receiving bank are different
	if receiverIndex == senderIndex {
		if receiverIndex < len(banks)-1 {
			receiverIndex++
		} else {
			receiverIndex = 0
		}
		// receiverIndex = rand.Intn(len(banks))
	}

	// create transaction and assign pointer
	result := &utils.Payment{
		Sender: utils.BankInfo{
			Name:    banks[senderIndex],
			Account: sendAcc,
		},
		Receiver: utils.BankInfo{
			Name:    banks[receiverIndex],
			Account: receiverAcc,
		},
		Amount: sum,
		Time:   time,
	}

	// format into payload for request
	payloadBuffer := new(bytes.Buffer)
	json.NewEncoder(payloadBuffer).Encode(result)
	return payloadBuffer
}
