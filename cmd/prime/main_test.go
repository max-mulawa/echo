package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	primeServer = "localhost"
)

var (
	ptrIsPrimeMethod = isPrimeMethod

	compositeNumber         = 123
	negativeCompositeNumber = -4
	negativeNonPrimeNumber  = -3
	primeNumer              = 13
	oneNumber               = 1
	zeroNumber              = 0
)

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func TestPrimeMethodServer(t *testing.T) {
	req := &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &primeNumer}
	payload := marshalRequest(t, req)

	expResponse := &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: true}
	respPayload := marshalResponse(t, expResponse)

	for _, tc := range []struct {
		descrition string
		payload    []byte
		clientsCnt uint
		sentCnt    uint
	}{
		{
			descrition: "single client and payload sent once",
			payload:    payload,
			clientsCnt: uint(1),
			sentCnt:    uint(1),
		},
		{
			descrition: "multi client and small payload sent once",
			payload:    payload,
			clientsCnt: uint(10),
			sentCnt:    uint(1),
		},
		{
			descrition: "multi client and multi payload sent once",
			payload:    payload,
			clientsCnt: uint(5),
			sentCnt:    uint(3),
		},
	} {
		t.Run(tc.descrition, func(t *testing.T) {
			wg := sync.WaitGroup{}
			wg.Add(int(tc.clientsCnt))
			for i := 1; i <= int(tc.clientsCnt); i++ {
				go func(index int) {
					defer wg.Done()

					conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", primeServer, serverPort))
					require.NoError(t, err)

					for sentCnt := 1; sentCnt <= int(tc.sentCnt); sentCnt++ {
						cntWrite, err := conn.Write(tc.payload)
						require.NoError(t, err)
						require.Equal(t, cntWrite, len(tc.payload))

						buff := make([]byte, 1024)
						cntRead, err := conn.Read(buff)
						require.NoError(t, err)
						require.Equal(t, respPayload, buff[0:cntRead])

						fmt.Printf("run %d completed (%d/%d) sent\n", index, sentCnt, tc.sentCnt)
					}

					fmt.Printf("run %d completed\n", index)
				}(i)
			}
			wg.Wait()
		})
	}
}

//Source: https://oeis.org/wiki/Nonprime_numbers
func TestPrimeChecks(t *testing.T) {
	for _, tc := range []struct {
		descrition  string
		request     *PrimeCheckRequest
		expResponse *PrimeCheckResponse
	}{
		{
			descrition:  "composite number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &compositeNumber},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "negative composite number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &negativeCompositeNumber},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &primeNumer},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: true},
		},
		{
			descrition:  "negative non-prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &negativeNonPrimeNumber},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "1 as non-prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &oneNumber},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "0 as non-prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &zeroNumber},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		//TODO: Any JSON number is a valid number, including floating-point values.
		// Note that non-integers can not be prime.
	} {
		t.Run(tc.descrition, func(t *testing.T) {
			reqPayload := marshalRequest(t, tc.request)
			respPayload := marshalResponse(t, tc.expResponse)

			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", primeServer, serverPort))
			require.NoError(t, err)

			_, err = conn.Write(reqPayload)
			require.NoError(t, err)

			buff := make([]byte, 1024)
			cntRead, err := conn.Read(buff)
			require.NoError(t, err)
			require.Equal(t, respPayload, buff[0:cntRead])
		})
	}
}

// handle multi request
func TestDummyData(t *testing.T) {
	payload := "\n"

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", primeServer, serverPort))
	require.NoError(t, err)

	_, err = conn.Write([]byte(payload))
	require.NoError(t, err)

	buff := make([]byte, 1024)
	cntRead, err := conn.Read(buff)
	require.NoError(t, err)
	fmt.Printf("read %d bytes", cntRead)
}

func TestMultiRequest(t *testing.T) {
	reqPayload := make([]byte, 0)
	reqPrime := &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &primeNumer}
	primeRequestPayload := marshalRequest(t, reqPrime)
	reqNonPrime := &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: &compositeNumber}
	nonPrimeRequestPayload := marshalRequest(t, reqNonPrime)
	for i := 0; i < 3; i++ {
		reqPayload = append(reqPayload, primeRequestPayload...)
		reqPayload = append(reqPayload, nonPrimeRequestPayload...)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", primeServer, serverPort))
	require.NoError(t, err)

	_, err = conn.Write(reqPayload)
	require.NoError(t, err)

	buff := make([]byte, 1024)
	cntRead, err := conn.Read(buff)
	require.NoError(t, err)

	respPayload := make([]byte, 0)
	respPrime := &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: true}
	primeRespPayload := marshalResponse(t, respPrime)
	respNonPrime := &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false}
	nonPrimeRespPayload := marshalResponse(t, respNonPrime)
	for i := 0; i < 3; i++ {
		respPayload = append(respPayload, primeRespPayload...)
		respPayload = append(respPayload, nonPrimeRespPayload...)
	}

	require.Equal(t, respPayload, buff[0:cntRead])
}

func TestPrimeInvalidRequestReplay(t *testing.T) {
	for _, tc := range []struct {
		descrition  string
		request     []byte
		expResponse []byte
	}{
		{
			descrition: "empty payload",
			request:    []byte("\n"),
		},
		{
			descrition: "empty json payload",
			request:    []byte("{}\n"),
		},
		{
			descrition: "invalid left bracket payload",
			request:    []byte("{\n"),
		},
		{
			descrition: "invalid right bracket payload",
			request:    []byte("}\n"),
		},
		{
			descrition: "text payload",
			request:    []byte("Ala ma kota\n"),
		},
		{
			descrition: "incorrect method payload",
			request:    []byte("{\"method\":\"isPrime2\",\"number\":7}\n"),
		},
		{
			descrition: "floating number payload",
			request:    []byte("{\"method\":\"isPrime\",\"number\":7.0}\n"),
		},
		{
			descrition: "invalid json payload",
			request:    []byte("{\"method\":\"isPrime\",\"number\":7.0\n"),
		},
		{
			descrition: "number sent as text payload",
			request:    []byte("{\"method\":\"isPrime\",\"number\":\"343453453\"}\n"),
		},
		{
			descrition: "number sent as array ",
			request:    []byte("{\"method\":\"isPrime\",\"number\":[1621288]}\n"),
		},
		{
			descrition: "number field is missing",
			request:    []byte("{\"method\":\"isPrime\"}\n"),
		},
		{
			descrition: "method field is missing",
			request:    []byte("{\"number\":1621288}\n"),
		},
		{
			descrition: "json payload with extra field",
			request:    []byte("{\"method\":\"isPrime\",\"number\":12343, \"text\":\"abcd efg hji\"}\n"),
		},
		{
			descrition: "big number",
			request:    []byte("{\"method\":\"isPrime\",\"number\":260940692798368945763214266270993929417123596856669561302575725}\n"),
		},
	} {
		t.Run(tc.descrition, func(t *testing.T) {
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", primeServer, serverPort))
			require.NoError(t, err)

			_, err = conn.Write(tc.request)
			require.NoError(t, err)

			buff := make([]byte, 1024)
			cntRead, err := conn.Read(buff)
			require.NoError(t, err)
			require.Equal(t, tc.request, buff[0:cntRead])
		})
	}
}

func marshalRequest(t *testing.T, req *PrimeCheckRequest) []byte {
	t.Helper()

	payload, err := json.Marshal(req)
	require.NoError(t, err)
	payload = append(payload, byte('\n'))
	return payload
}

func marshalResponse(t *testing.T, response *PrimeCheckResponse) []byte {
	respPayload, err := json.Marshal(response)
	require.NoError(t, err)

	respPayload = append(respPayload, '\n')

	return respPayload
}
