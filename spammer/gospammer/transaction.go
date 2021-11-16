package gospammer

import (
	"fmt"
	"time"
	"bytes"
	"bufio"
	"net/http"
	"math/rand"
	"encoding/json"
)

const (
	lo = 0
	hi = 1000

	channelCapacity = 100000
	contentType = "transaction/json"
	timeout = 3 * time.Second
)

// transaction stores information constituting a transaction.
type transaction struct {
	Sender string
	Receiver string
	Sum int
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

// TransactionSpammer sends *times POST requests concurrently,
// to the specified dest. The goal is to send requests as 
// close to simultaneous as possible.
func TransactionSpammer(dest string, times int) {

	reqCollection := make([]bytes.Buffer, times)
	respChannel := make(chan *http.Response, channelCapacity)

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

	// Send the number of requests
	fmt.Printf("\nSending %d requests concurrently.\n", len(reqCollection))
	for i := 0; i < times; i++ {
		go SendTransaction(dest, &reqCollection[i], respChannel)
	}
	
	// Process the responses
	for i := 0; i < times; i++ {
		select {
		case resp := <- respChannel:
			ProcessResponse(resp)
		case <- time.After(timeout):
			fmt.Println("Waited response for too long.")
			break
		}
	}
	
}

// SendTransaction sends a post request with transaction information
// and prints the response's contents.
func SendTransaction(dest string, payload *bytes.Buffer, respChannel chan<- *http.Response) {

	// Post request 
	resp, err := http.Post(dest, contentType, payload)
	if err != nil {
		panic(err)
	}

	// Pass response in the channel
	respChannel <- resp
}

// Prints the contents of responses and "handles" errors.
func ProcessResponse(resp *http.Response) {
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
			receiverIndex=0 
		}
		// receiverIndex = rand.Intn(len(banks))
	}

	// create transaction and assigh pointer
	result := &transaction{
		Sender: banks[senderIndex],
		Receiver: banks[receiverIndex],
		Sum: sum,
	}

	// format into payload for request
	payloadBuffer := new(bytes.Buffer)
	json.NewEncoder(payloadBuffer).Encode(result)
	return payloadBuffer
}