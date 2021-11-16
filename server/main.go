package main 

import (
	"fmt"
	"time"
	"net/http"

	"github.com/IsakJones/polka/server/views"
)

func main() {
	
	http.HandleFunc("/transaction", views.Transaction)
	http.HandleFunc("/hello", views.Hello)

	// hostName, err := os.Hostname()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(hostName)

	go	http.ListenAndServe(":8090", nil)

	fmt.Println("Requests received:")
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		views.PrintCounter()
	}
}
