package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"
)

const (
	serverPort = 8887

	// destinationHost = "chat.protohackers.com"
	// destinationPort = 16963

	destinationHost = "localhost"
	destinationPort = 8888
)

func main() {
	startServer()
}

func startServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		fmt.Printf("failed to listen on port %d port: %v", serverPort, err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		conn.SetDeadline(time.Now().Add(time.Second * 120 * 10)) //TODO: remove 10
		go handleConnection(conn)
	}
}

func handleConnection(sconn net.Conn) {
	defer func() {
		fmt.Print("Closing source connection on proxy\n")
		sconn.Close()
	}()

	request := make(chan []byte)
	response := make(chan []byte)

	dconn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", destinationHost, destinationPort))
	if err != nil {
		fmt.Printf("failed to connect to `%s:%d`: %v", destinationHost, destinationPort, err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		fmt.Println("Closing destination connection on proxy")
		dconn.Close()
		cancel()
	}()

	go ReadRequest(sconn, request, ctx, cancel)
	go WriteRequest(dconn, request, ctx, cancel)
	go ReadResponse(dconn, response, ctx, cancel)
	WriteResponse(sconn, response, ctx, cancel)
}

func WriteResponse(sconn net.Conn, response <-chan []byte, ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-response:
			_, err := sconn.Write(resp)
			if err != nil {
				fmt.Println("failed on writing to source", err)
				cancel()
				return
			}
		}
	}
}

func ReadRequest(conn net.Conn, request chan<- []byte, ctx context.Context, cancel context.CancelFunc) {
	r := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := r.ReadBytes('\n')
			if err != nil {
				fmt.Println("failed on reading from source", err)
				cancel()
				return
			}
			request <- req
		}
	}
}

func ReadResponse(conn net.Conn, response chan<- []byte, ctx context.Context, cancel context.CancelFunc) {
	r := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			resp, err := r.ReadBytes('\n')
			if err != nil {
				fmt.Println("failed on read from destination", err)
				cancel()
				return
			}
			response <- resp
		}

	}
}

func WriteRequest(conn net.Conn, request <-chan []byte, ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-request:
			_, err := conn.Write(req)
			if err != nil {
				fmt.Println("failed on write to destination", err)
				cancel()
				return
			}
		}
	}
}

var (
	bogusCoinExp = regexp.MustCompile("(7[0-9a-zA-Z]{25,34})")
)

func rewrite(msg []byte) []byte {
	txt := string(msg)
	bogusCoinExp.FindStringSubmatchIndex(txt)

}
