package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/fxtlabs/primes"
)

const (
	serverPort = 8888

	isPrimeMethod = "isPrime"
)

func main() {
	startServer()
}

func startServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		conn.SetDeadline(time.Now().Add(time.Second * 120))
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Print("Closing connection on serer\n")
		conn.Close()
	}()
	bufSize := 1024
	totalCount := 0
	payload := make([]byte, 0, bufSize)
	buffer := make([]byte, bufSize)
	for {
		count, err := conn.Read(buffer)

		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error - %s\n", err)
			} else {
				fmt.Print("Client closed connection\n")
			}
			break
		}

		if count > 0 {
			totalCount += count
			payload = append(payload, buffer[0:count]...)

			if payload[len(payload)-1] == byte('\n') {
				//end of payload
				requestsPayloads := bytes.Split(payload, []byte("\n"))
				//remove the last one empty
				requestsPayloads = requestsPayloads[:(len(requestsPayloads) - 1)]
				respPayloads := make([][]byte, 0)

				for _, reqPayload := range requestsPayloads {
					request := &PrimeCheckRequest{}
					if err := json.Unmarshal(reqPayload, request); err != nil {
						fmt.Printf("failed to unmarshal payload request: %v\n", err)
						//conn.Write(payload)
						//return
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					if !(request.Method == isPrimeMethod) {
						fmt.Printf("method not supported: %s\n", request.Method)
						//conn.Write(payload)
						//return
						respPayloads = append(respPayloads, reqPayload)
						continue
					} else {
						isPrimeNumber := primes.IsPrime(request.Number)
						response := &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: isPrimeNumber}
						jsonResponse, err := json.Marshal(response)
						if err != nil {
							fmt.Printf("failed serializing response %v: %v\n", response, err)
						}
						//jsonResponse = append(jsonResponse, '\n')
						//conn.Write(jsonResponse)
						respPayloads = append(respPayloads, jsonResponse)
					}
				}
				payload = make([]byte, 0, bufSize)

				respPayload := bytes.Join(respPayloads, []byte{'\n'})
				respPayload = append(respPayload, byte('\n'))
				conn.Write(respPayload)
			}
		}
	}
}

type PrimeCheckRequest struct {
	Method string `json:"method"`
	Number int    `json:"number"`
}

type PrimeCheckResponse struct {
	Method  string `json:"method"`
	IsPrime bool   `json:"prime"`
}
