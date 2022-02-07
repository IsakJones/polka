package gospammer

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	lo = 0
	hi = 1000

	contentType = "transaction/json"
	timeout     = 3 * time.Second
)

var badResponses *uint32

// transaction stores information constituting a transaction.
type transaction struct {
	Sender   bankInfo
	Receiver bankInfo
	Amount   int
	Time     time.Time
}

type bankInfo struct {
	Name    string
	Account int
}

// list of the 10 largest US banks
var banks = []string{
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

var c *http.Client

// TransactionSpammer sends transactionNumber POST requests concurrently,
// to the specified dest. The goal is to send requests as
// close to simultaneous as possible. Returns number of bad requests.
func TransactionSpammer(dest string, maxGoroutines, transactionNumber uint) uint32 {

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
		go Worker(lo, hi, dest, workChan, doneChanNew)
	}

	// Assign work to workers
	log.Printf("filling work channel")
	for i := uint(0); i < transactionNumber; i++ {
		workChan <- true
		// if i == maxSubscriberGoroutines {
		// 	log.Printf("Reached %d", maxSubscriberGoroutines)
		// }
	}

	// End workers
	log.Printf("killing workers")
	for i := uint(0); i < maxGoroutines; i++ {
		doneChanNew <- true
	}
	log.Printf("Finished.")

	return *badResponses
}

func Worker(lo, hi int, dest string, work <-chan interface{}, done <-chan interface{}) {
	for {
		select {
		case <-done:
			return
		case <-work:
			payload := generateTransaction(lo, hi)

			// Post request
			// start := time.Now()
			sendTransaction(dest, payload)
			// end := time.Now()
			// log.Printf("Sending the transation took %s", end.Sub(start))
		}
	}
}

// sendTransaction sends a POST request with a json-encoded
// payment to the load balancer.
func sendTransaction(dest string, payload *bytes.Buffer) {
	resp, err := c.Post(dest, contentType, payload)
	// req, err := http.NewRequest(http.MethodPost, dest, payload)
	// if err != nil {
	// 	log.Printf("Error making request: %s", err.Error())
	// }
	// req.Close = true
	// resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error sending request: %s", err.Error())
		return
	}
	defer resp.Body.Close()
	// In case of failure, print
	if resp.StatusCode > 299 {
		atomic.AddUint32(badResponses, 1)
		log.Printf("Received response with bad status code: %s", resp.Body)
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
	result := &transaction{
		Sender: bankInfo{
			Name:    banks[senderIndex],
			Account: sendAcc,
		},
		Receiver: bankInfo{
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
