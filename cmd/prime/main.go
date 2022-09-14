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

	isPrimeMethod string = "isPrime"
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
		fmt.Print("Closing connection on server\n")
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

					var requestFields map[string]interface{}
					if err := json.Unmarshal(reqPayload, &requestFields); err != nil {
						fmt.Printf("failed to unmarshal to a map payload request: %v\n", err)
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					var allowedFields = map[string]bool{
						"method": true,
						"number": true,
					}

					superfluousField := ""
					for fieldName, _ := range requestFields {
						if ok := allowedFields[fieldName]; !ok {
							superfluousField = fieldName
							break
						}
					}
					if superfluousField != "" {
						fmt.Printf("failed to unmarshal payload request, extra field present: %s\n", superfluousField)
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					request := &PrimeCheckRequest{}
					if err := json.Unmarshal(reqPayload, request); err != nil {
						fmt.Printf("failed to unmarshal payload request: %v\n", err)
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					if request.Method == nil {
						fmt.Print("method field is missing")
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					if request.Number == nil {
						fmt.Print("number field is missing")
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					if !(*request.Method == isPrimeMethod) {
						fmt.Printf("method not supported: %s\n", *request.Method)
						respPayloads = append(respPayloads, reqPayload)
						continue
					} else {

						isPrimeNumber := primes.IsPrime(*request.Number)
						response := &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: isPrimeNumber}
						jsonResponse, err := json.Marshal(response)
						if err != nil {
							fmt.Printf("failed serializing response %v: %v\n", response, err)
						}
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
	Method *string `json:"method"`
	Number *int    `json:"number"`
}

type PrimeCheckResponse struct {
	Method  string `json:"method"`
	IsPrime bool   `json:"prime"`
}
