package views

import (
	"fmt"
	// "bufio"
	"net/http"
	"encoding/json"
)

// Transaction captures the data in transactions POST requests.
type transaction struct {
	Sender string
	Receiver string
	Sum int
}

// Hello verifies that get requests work.
func Hello(writer http.ResponseWriter, req *http.Request) {

	fmt.Println("Successfully got a hello HTTP request!")
	fmt.Fprintf(writer, "Hello!!!\n")
}

func Transaction(writer http.ResponseWriter, req *http.Request) {

	// Verify that a transaction has been received
	fmt.Println("Successfully received a transaction!")
	if req.Method == "POST" {
		fmt.Fprintf(writer, "Successfully sent a post!\n")
	} else {
		fmt.Fprintf(writer, "Something wrong happened...")
	}
	//fmt.Println(*req.Body)

	// Read the body
	var current transaction
	err := json.NewDecoder(req.Body).Decode(&current)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(writer, "Received transaction: %+v\n", current)
	fmt.Printf("Received transaction: %+v\n", current)

}

// // readTransaction prints out the contents of the json
// func readTransaction(req  *http.Request) {
// 	fmt.Printf("%t", req.)

// }