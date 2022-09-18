package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/fxtlabs/primes"
)

const (
	serverPort = 8888

	isPrimeMethod string = "isPrime"
)

// https://oeis.org/wiki/Nonprime_numbers
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

					allowedFieldsCnt := 0
					for fieldName := range requestFields {
						if ok := allowedFields[fieldName]; ok {
							allowedFieldsCnt++
						}
					}

					if len(allowedFields) != allowedFieldsCnt {
						fmt.Printf("required fields are missing")
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					switch method := requestFields["method"].(type) {
					case string:
						if method != isPrimeMethod {
							fmt.Printf("method not supported: %s\n", method)
							respPayloads = append(respPayloads, reqPayload)
							continue
						}
					default:
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					request := &PrimeCheckRequest{}
					if err := json.Unmarshal(reqPayload, request); err != nil {
						fmt.Printf("failed to unmarshal payload request: %v\n", err)
						respPayloads = append(respPayloads, reqPayload)
						continue
					}

					isPrimeNumber := false

					switch request.Number.valueType {
					case Integer:
						isPrimeNumber = primes.IsPrime(request.Number.value.(int))
					case Float:
						isPrimeNumber = false
					case BigInt:
						bigNumber := request.Number.value.(big.Int)
						isPrimeNumber = bigNumber.ProbablyPrime(0)
					}

					response := &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: isPrimeNumber}
					jsonResponse, err := json.Marshal(response)
					if err != nil {
						fmt.Printf("failed serializing response %v: %v\n", response, err)
					}
					respPayloads = append(respPayloads, jsonResponse)
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
	Method *string    `json:"method"`
	Number NumberInfo `json:"number"`
}

type NumberInfo struct {
	value     interface{}
	valueType ValueType
}

func (n NumberInfo) MarshalJSON() ([]byte, error) {
	return []byte(n.value.(string)), nil
}

func (n *NumberInfo) UnmarshalJSON(b []byte) error {
	val := string(b)
	n.value = val
	valInteger, err := strconv.ParseInt(val, 10, 32)
	if err == nil {
		n.valueType = Integer
		n.value = int(valInteger)
		return nil
	} else if numErr, ok := err.(*strconv.NumError); ok {
		if numErr.Err == strconv.ErrRange {
			var bigIntVal big.Int
			_, ok := bigIntVal.SetString(val, 10)
			if ok {
				n.valueType = BigInt
				n.value = bigIntVal
				return nil
			}
		}
	}

	floatVal, err := strconv.ParseFloat(val, 64)
	if err == nil {
		n.valueType = Float
		n.value = floatVal
		return nil
	}

	return fmt.Errorf("failed to convert number to numerical value %s", val)
}

type ValueType string

var (
	Integer ValueType = "Integer"
	Float   ValueType = "Float"
	BigInt  ValueType = "BitInt"
)

type PrimeCheckResponse struct {
	Method  string `json:"method"`
	IsPrime bool   `json:"prime"`
}
