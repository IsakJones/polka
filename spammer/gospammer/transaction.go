package gospammer

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	lo = 0
	hi = 1000

	maxSubscriberGoroutines = 7000
	contentType             = "transaction/json"
	timeout                 = 3 * time.Second
)

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

// TransactionSpammer sends transactionNumber POST requests concurrently,
// to the specified dest. The goal is to send requests as
// close to simultaneous as possible.
func TransactionSpammer(dest string, transactionNumber int) {

	workChan := make(chan interface{})
	doneChanNew := make(chan interface{})

	log.Printf("initializing workers")
	for i := 0; i < maxSubscriberGoroutines; i++ {
		go Worker(lo, hi, dest, workChan, doneChanNew)
	}

	log.Printf("filling work channel")
	for i := 0; i < transactionNumber; i++ {
		workChan <- true
		if i == maxSubscriberGoroutines {
			log.Printf("Reached %d", maxSubscriberGoroutines)
		}
	}

	log.Printf("killing workers")
	// Tell workers to end
	for i := 0; i < maxSubscriberGoroutines; i++ {
		doneChanNew <- true
	}
	log.Printf("Finished.")

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

func sendTransaction(dest string, payload *bytes.Buffer) {
	resp, err := http.Post(dest, contentType, payload)
	if err != nil {
		log.Printf("Error sending request: %s", err)
		return
	}
	defer resp.Body.Close()
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
