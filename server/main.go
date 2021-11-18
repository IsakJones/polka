package main 

import (
	"fmt"
	"time"
	"net/http"

	"github.com/IsakJones/polka/server/transactions"
)

const (
	transPath = "/transaction"
	helloPath = "/hello"
	port = "8090"
)


func main() {
	
	// http.HandleFunc("/transaction", transactions.Manage)
	// http.HandleFunc("/hello", transactions.Hello)

	// hostName, err := os.Hostname()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(hostName)
	smux := http.NewServeMux()

	smux.HandleFunc(transPath, transactions.Manage())

	go	smux.ListenAndServe(port, nil)

	fmt.Println("Requests received:")
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		transactions.PrintCounter()
		transactions.PrintDues()
	}
}
