package main

import (
	"log"
	"time"
)

func liltest() {

	channel := make(chan bool)

	go func() {
		time.Sleep(time.Second)
		channel <- true
	}()

	for {
		select {
		case <-channel:
			log.Println("Got signal to end...")
			break
		default:
			log.Println("waiting...")
			time.Sleep(100 * time.Millisecond)
		}
	}
}
