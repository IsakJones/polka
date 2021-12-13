package gospammer

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	lo = 0
	hi = 1000

	contentType = "transaction/json"
	timeout     = 3 * time.Second
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

	/*

		doneChan := make(chan bool)
		codeChan := make(chan string)
	*/

	/*
			 - Subscriber
		==== - Subscriber
			 ....
	*/

	//var wg sync.WaitGroup
	workChan := make(chan interface{})
	doneChanNew := make(chan interface{})
	maxSubscriberGoroutines := 1000
	//maxPublisherGoroutines := 100

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

	// Spam into work for n requests
	/*
		for i := 0; i < maxPublisherGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < transactionNumber/maxPublisherGoroutines; j++ {
					workChan <- true
				}
			}()
		}
		wg.Wait()
	*/

	log.Printf("killing workers")
	// Tell workers to end
	for i := 0; i < maxSubscriberGoroutines; i++ {
		doneChanNew <- true
	}
	log.Printf("All has ended.\n")

	/*
		// The main body is off in a goroutine...
		go func() {
			// Send requests
			rand.Seed(time.Now().UnixNano())
			log.Printf("Sending %d requests concurrently.\n", transactionNumber)
			for i := 0; i < transactionNumber; i++ {
				wg.Add(1)
				go SendTransaction(lo, hi, dest, codeChan, &wg)
			}
			wg.Wait()
			doneChan <- true
		}()

		// So it's easier to detect whether anything goes wrong.
		noErrors := true
		select {
		case respStatus := <-codeChan:
			noErrors = false
			log.Printf("Request failed: %s\n", respStatus)
		case <-doneChan:
			log.Printf("Completed all request attempts.\n")
			break
		}

		if noErrors {
			log.Printf("All requests sent successfully!")
		}
	*/
}

func Worker(lo, hi int, dest string, work <-chan interface{}, done <-chan interface{}) {
	for {
		select {
		case <-done:
			return
		case <-work:
			SendTransactionNew(lo, hi, dest)
		}
	}
}

func SendTransactionNew(lo, hi int, dest string) {
	payload := GenerateTransaction(lo, hi)
	// Set timeout
	/*
		client := http.Client{
			Timeout: timeout * time.Second,
		}
	*/
	// Post request
	// start := time.Now()
	ActuallySendTransaction(dest, payload)
	// end := time.Now()
	// log.Printf("Actually send transation took %s", end.Sub(start))
}

func ActuallySendTransaction(dest string, payload *bytes.Buffer) {
	resp, err := http.Post(dest, contentType, payload)
	if err != nil {
		log.Printf("Error sending request: %s", err)
		return
	}
	defer resp.Body.Close()
}

// SendTransaction sends a post request with transaction information
// and prints the response's contents.
func SendTransaction(lo, hi int, dest string, codeChan chan<- string, wg *sync.WaitGroup) {

	// Report to waitgroup
	defer wg.Done()

	// Generate random transaction
	payload := GenerateTransaction(lo, hi)

	// Set timeout
	client := http.Client{
		Timeout: timeout * time.Second,
	}
	// Post request
	resp, err := client.Post(dest, contentType, payload)
	if err != nil {
		log.Printf("Error sending request: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		codeChan <- resp.Status
	}
}

// Returns a buffer fit for the POST request encoding
// a transaction struct.
func GenerateTransaction(lo, hi int) *bytes.Buffer {

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

// For testing purposes
// func GenerteTransaction() *bytes.Buffer {

// 	// create transaction and assigh pointer
// 	result := &transaction{
// 		Sender:   "JPMorgan",
// 		Receiver: "BofA",
// 		Amount:   100,
// 	}

// 	// format into payload for request
// 	payloadBuffer := new(bytes.Buffer)
// 	json.NewEncoder(payloadBuffer).Encode(result)
// 	return payloadBuffer
// }
