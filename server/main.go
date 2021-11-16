package main 

import (
	// "fmt"
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

	http.ListenAndServe(":8090", nil)
}
