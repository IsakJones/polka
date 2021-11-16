package transactions

import (
	"fmt"
	"time"
	// "bufio"
	"net/http"
	"sync/atomic"
	"encoding/json"
)

// Transaction captures the data in transactions POST requests.
type Transaction struct {
	Sender string
	Receiver string
	Sum int
}

var counter uint64
var transactionChannel = make(chan *Transaction, 500)

// Hello verifies that get requests work.
func Hello(writer http.ResponseWriter, req *http.Request) {

	fmt.Println("Successfully got a hello HTTP request!")
	fmt.Fprintf(writer, "Hello!!!\n")
}

func Transaction(writer http.ResponseWriter, req *http.Request) {

	// Verify that a transaction has been received
	// fmt.Println("Successfully received a transaction!")
	// if req.Method == "POST" {
	// 	fmt.Fprintf(writer, "Successfully sent a post!\n")
	// } else {
	// 	fmt.Fprintf(writer, "Something wrong happened...")
	// }
	//fmt.Println(*req.Body)

	// Read the body
	var current transaction
	err := json.NewDecoder(req.Body).Decode(&current)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// Clear transaction and send response
	clearTransaction(&current)
	fmt.Fprintf(writer, "Received transaction: %+v\n", current)
	// fmt.Printf("Received transaction: %+v\n", current)
	atomic.AddUint64(&counter, 1)
}

func PrintCounter() {
	fmt.Printf("Processed %d transactions by %s.\n", counter, time.Now().Format("15:04:05"))
}

// // readTransaction prints out the contents of the json
// func readTransaction(req  *http.Request) {
// 	fmt.Printf("%t", req.)

// }