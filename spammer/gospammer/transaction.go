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

	var wg sync.WaitGroup

	doneChan := make(chan bool)
	codeChan := make(chan string)

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
}

// SendTransaction sends a post request with transaction information
// and prints the response's contents.
func SendTransaction(lo, hi int, dest string, codeChan chan<- string, wg *sync.WaitGroup) {

	// Report to waitgroup
	defer wg.Done()

	// Generate random transaction
	payload := GenerateTransaction(lo, hi)

	// Post request
	resp, err := http.Post(dest, contentType, payload)
	if err != nil {
		log.Printf("Error sending request: %s", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		codeChan <- resp.Status
	}

	// // Scan the body and check for errors
	// scanner := bufio.NewScanner(resp.Body)
	// for scanner.Scan() {
	// 	log.Println(scanner.Text())
	// }
	// if err := scanner.Err(); err != nil {
	// 	return err
	// }
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
