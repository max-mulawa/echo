package main

import (
	"encoding/json"
	"fmt"
	"math"
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

	compositeNumber         int = 123
	negativeCompositeNumber int = -4
	negativeNonPrimeNumber  int = -3
	primeNumer              int = 13
	oneNumber               int = 1
	zeroNumber              int = 0
)

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func getIntNumber(n int) NumberInfo {
	return NumberInfo{
		value: fmt.Sprintf("%d", n),
	}
}

func TestPrimeMethodServer(t *testing.T) {
	req := &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(primeNumer)}
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
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(compositeNumber)},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "negative composite number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(negativeCompositeNumber)},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(primeNumer)},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: true},
		},
		{
			descrition:  "negative non-prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(negativeNonPrimeNumber)},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "1 as non-prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(oneNumber)},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "0 as non-prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(zeroNumber)},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "big int as prime number",
			request:     &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: NumberInfo{value: "20988936657440586486151264256610222593863921"}},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: true},
		},
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

func TestNumbersAreNotPrime(t *testing.T) {
	for _, tc := range []struct {
		description string
		number      string
	}{
		{
			description: "value with dot",
			number:      "7000.33",
		},
		{
			description: "value with comma",
			number:      "700033.00",
		},
		{
			description: "bigint that is not prime",
			number:      "260940692798368945763214266270993929417123596856669561302575725",
		},
		{
			description: "max int64 that is not prime",
			number:      fmt.Sprintf("%d", math.MaxInt64),
		},
		{
			description: "max float 64 that is not prime",
			number:      fmt.Sprintf("%f", math.MaxFloat64),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			payload := fmt.Sprintf("{\"method\":\"isPrime\",\"number\":%s, \"number2\":\"ok\"}\n", tc.number)

			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", primeServer, serverPort))
			require.NoError(t, err)

			_, err = conn.Write([]byte(payload))
			require.NoError(t, err)

			buff := make([]byte, 1024)
			cntRead, err := conn.Read(buff)
			require.NoError(t, err)
			fmt.Printf("read %d bytes", cntRead)
			respNonPrime := &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false}
			nonPrimeRespPayload := marshalResponse(t, respNonPrime)

			require.Equal(t, nonPrimeRespPayload, buff[:cntRead])
		})
	}
}

func TestMultiRequest(t *testing.T) {
	reqPayload := make([]byte, 0)
	reqPrime := &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(primeNumer)}
	primeRequestPayload := marshalRequest(t, reqPrime)
	reqNonPrime := &PrimeCheckRequest{Method: &ptrIsPrimeMethod, Number: getIntNumber(compositeNumber)}
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
