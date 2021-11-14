package main 

import (
	// "os"
	"fmt"
	"net/http"
)

func main() {
	
	http.HandleFunc("/transaction", transaction)
	http.HandleFunc("/hello", hello)

	// hostName, err := os.Hostname()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(hostName)

	http.ListenAndServe(":8090", nil)
}

func hello(writer http.ResponseWriter, req *http.Request) {

	fmt.Println("Successfully got a hello HTTP request!")
	fmt.Println(req)
	fmt.Fprintf(writer, "Hello!!!\n")
}

func transaction(writer http.ResponseWriter, req *http.Request) {

	fmt.Println("Successfully received a transaction!")
	if req.Method == "POST" {
		fmt.Fprintf(writer, "Successfully sent a post!")
	} else {
		fmt.Fprintf(writer, "Something wrong happened...")
	}
	//fmt.Println(*req.Body)
}