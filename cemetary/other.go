package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type Server struct {
	listener net.Listener
	quit     chan bool
	exited   chan bool
}

func NewServer() *Server {
	addr, err := net.ResolveTCPAddr("tcp4", ":9999")
	// TODO: return nil, error and decide how to handle it in the calling function
	if err != nil {
		fmt.Println("Failed to resolve address", err.Error())
		os.Exit(1)
	}

	// TODO: return nil, error and decide how to handle it in the calling function
	listener, err := net.Listen("tcp", addr.String())
	if err != nil {
		fmt.Println("Failed to create listener", err.Error())
		os.Exit(1)
	}

	// TODO: do not use this syntax, add the field names
	srv := &Server{
		listener,
		make(chan bool),
		make(chan bool),
	}
	// TODO: no need to export Serve as it is only called internally
	go srv.Serve()
	return srv
}

func (srv *Server) Serve() {
	var handlers sync.WaitGroup
	for {
		select {
		case <-srv.quit:
			fmt.Println("Shutting down...")
			srv.listener.Close()
			handlers.Wait()
			close(srv.exited)
			return
		default:
			//fmt.Println("Listening for clients")
			srv.listener.SetDeadline(time.Now().Add(1e9))
			conn, err := srv.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				fmt.Println("Failed to accept connection:", err.Error())
			}
			handlers.Add(1)
			go func() {
				// FIXME: handle returned error here (just log it)
				// FIXME: determine ID (why?)
				srv.handleConnection(conn, 0)
				handlers.Done()
			}()
		}
	}
}

func (srv *Server) handleConnection(conn net.Conn, id int) error {
	fmt.Println("Accepted connection from", conn.RemoteAddr())

	defer func() {
		fmt.Println("Closing connection from", conn.RemoteAddr())
		conn.Close()
	}()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read error", err.Error())
		return err
	}
	return nil
}

func (srv *Server) Stop() {
	fmt.Println("Stop requested")
	// XXX: You cannot use the same channel in two directions.
	//      The order of operations on the channel is undefined.
	close(srv.quit)
	<-srv.exited
	fmt.Println("Stopped successfully")
}

// func main() {
// 	srv := NewServer()
// 	time.Sleep(2 * time.Second)
// 	srv.Stop()
// }
