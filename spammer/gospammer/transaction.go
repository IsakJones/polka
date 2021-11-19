package gospammer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
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
	Sender   string
	Receiver string
	Amount   int
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

// TransactionSpammer sends #times POST requests concurrently,
// to the specified dest. The goal is to send requests as
// close to simultaneous as possible.
func TransactionSpammer(dest string, times int) {

	// This splice will store the generated transactions
	reqCollection := make([]bytes.Buffer, times)

	// Generate request payloads
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < times; i++ {
		reqCollection[i] = *GenerateTransaction(lo, hi)
	}

	// Ask for confirmation
	var answer string
	for {
		fmt.Printf("%d payloads on standby. Send? (yes/no)\t", len(reqCollection))
		fmt.Scanf("%s", &answer)

		if answer == "yes" {
			break
		} else if answer == "no" {
			fmt.Println("Payloads suspended.")
			return
		} else {
			fmt.Println("Answer not recognized, please try again.")
		}
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	// Send requests
	fmt.Printf("\nSending %d requests concurrently.\n", len(reqCollection))
	for i := 0; i < times; i++ {
		go SendTransaction(dest, &reqCollection[i], &wg)
	}

}

// SendTransaction sends a post request with transaction information
// and prints the response's contents.
func SendTransaction(dest string, payload *bytes.Buffer, wg *sync.WaitGroup) {

	// Update waitgroup
	wg.Add(1)
	defer wg.Done()

	// Post request
	resp, err := http.Post(dest, contentType, payload)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)

	// Scan the body and check for errors
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

// Returns a buffer fit for the POST request encoding
// a transaction struct.
func GenerateTransaction(lo, hi int) *bytes.Buffer {

	// calculate basic transaction attributes
	sum := rand.Intn(hi-lo) + lo
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

	// create transaction and assigh pointer
	result := &transaction{
		Sender:   banks[senderIndex],
		Receiver: banks[receiverIndex],
		Amount:   sum,
	}

	// format into payload for request
	payloadBuffer := new(bytes.Buffer)
	json.NewEncoder(payloadBuffer).Encode(result)
	return payloadBuffer
}

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