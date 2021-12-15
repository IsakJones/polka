package handlers

import (
	"fmt"
	"net/http"
)

// Hello verifies that get requests work.
func Hello(writer http.ResponseWriter, req *http.Request) {

	fmt.Println("Successfully got a hello HTTP request!")
	fmt.Fprintf(writer, "Hello!!!\n")
}
