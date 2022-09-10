package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

const (
	echoPort = 7777
)

func main() {
	startServer()
}

func startServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", echoPort))
	if err != nil {
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	bufSize := 1024
	buffer := make([]byte, bufSize)
	for {
		count, err := conn.Read(buffer)
		if count > 0 {
			conn.Write(buffer[0:count])
		}
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error - %s\n", err)
			} else {
				fmt.Print("Closing connection\n")
			}
			break
		}
	}
}
