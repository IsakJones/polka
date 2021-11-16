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

// SendTransaction sends a post request with transaction information
// and prints the response's contents.
func SendTransaction(dest string) {

	contentType := "transaction/json"

	// Generate transaction and encode to json
	transaction := generateTransaction(0, 1000)
	fmt.Printf("Generated transaction %+v\n", transaction)
	payloadBuffer := new(bytes.Buffer)
	json.NewEncoder(payloadBuffer).Encode(transaction)
	
	// Post request 
	resp, err := http.Post(dest, contentType, payloadBuffer)
	if err != nil {
		panic(err)
	}

	// Start scanning the body
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	// Print transaction body
	fmt.Println("Response status:", resp.Status)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

}

func generateTransaction(lo, hi int) *transaction {


	// TODO find way to only init once
	rand.Seed(time.Now().UnixNano())

	// calculate basic transaction attributes
	sum := rand.Intn(hi-lo) + lo
	senderIndex := rand.Intn(len(banks))
	receiverIndex := rand.Intn(len(banks))
	if receiverIndex == senderIndex {
		if receiverIndex < len(banks)-1 { 
			receiverIndex++ 
		} else {
			receiverIndex=0 
		}
		// receiverIndex = rand.Intn(len(banks))
	}

	// return transaction pointer
	return &transaction{
		Sender: banks[senderIndex],
		Receiver: banks[receiverIndex],
		Sum: sum,
	}
}