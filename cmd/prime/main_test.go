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

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func TestPrimeMethodServer(t *testing.T) {
	req := &PrimeCheckRequest{Method: isPrimeMethod, Number: 17}
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
			request:     &PrimeCheckRequest{Method: isPrimeMethod, Number: 123},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "negative composite number",
			request:     &PrimeCheckRequest{Method: isPrimeMethod, Number: -4},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "prime number",
			request:     &PrimeCheckRequest{Method: isPrimeMethod, Number: 13},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: true},
		},
		{
			descrition:  "negative non-prime number",
			request:     &PrimeCheckRequest{Method: isPrimeMethod, Number: -3},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "1 as non-prime number",
			request:     &PrimeCheckRequest{Method: isPrimeMethod, Number: 1},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		{
			descrition:  "0 as non-prime number",
			request:     &PrimeCheckRequest{Method: isPrimeMethod, Number: 0},
			expResponse: &PrimeCheckResponse{Method: isPrimeMethod, IsPrime: false},
		},
		//TODO: Any JSON number is a valid number, including floating-point values.
		// Note that non-integers can not be prime.

	} {
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
	}
}

//TODO:
// handle multi request
// {"number":99198753,"method":"isPrime"}
// {"number":84263327,"method":"isPrime"}
// {"method":"isPrime","number":73521363}
// {"method":"isPrime","number":74683887}
// {"number":99652937,"method":"isPrime"}
// {"method":"isPrime","number":89125073}
// {"number":52593379,"method":"isPrime"}
// {"method":"isPrime","number":2896170}
// {"method":"isPrime","number":20211539}
// {"method":"isPrime","number":10876997}
// {"number":80035997,"method":"isPrime"}
// {"method":"isPrime","number":35170376}
// {"number":81918192,"method":"isPrime"}
// {"method":"isPrime","number":48176357}
// {"method":"isPrime","number":66536893}
// {"number":64228979,"method":"isPrime"}
// {"method":"isPrime","number":31406846}

func TestPrimeInvalidRequestReplay(t *testing.T) {
	for _, tc := range []struct {
		descrition  string
		request     []byte
		expResponse []byte
	}{
		{
			descrition: "empty payload replayed",
			request:    []byte("{}\n"),
		},
		{
			descrition: "incorrect method payload replayed",
			request:    []byte("{\"method\":\"isPrime2\",\"number\":7}\n"),
		},
		{
			descrition: "floating number payload replayed",
			request:    []byte("{\"method\":\"isPrime\",\"number\":7.0}\n"),
		},
		{
			descrition: "invalid json payload replayed",
			request:    []byte("{\"method\":\"isPrime\",\"number\":7.0\n"),
		},
		{
			descrition: "invalid number format payload replayed",
			request:    []byte("{\"method\":\"isPrime\",\"number\":\"343453453\"}\n"),
		},
	} {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", primeServer, serverPort))
		require.NoError(t, err)

		_, err = conn.Write(tc.request)
		require.NoError(t, err)

		buff := make([]byte, 1024)
		cntRead, err := conn.Read(buff)
		require.NoError(t, err)
		require.Equal(t, tc.request, buff[0:cntRead])
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
