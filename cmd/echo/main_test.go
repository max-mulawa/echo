package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func TestTCPEchoServer(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		descrition string
		payload    []byte
		clientsCnt uint
		sentCnt    uint
	}{
		{
			descrition: "single client and small payload sent once",
			payload:    []byte("abc"),
			clientsCnt: uint(1),
			sentCnt:    uint(1),
		},
		{
			descrition: "multi client and small payload sent once",
			payload:    []byte("abc"),
			clientsCnt: uint(10),
			sentCnt:    uint(1),
		},
		{
			descrition: "multi client and multi payload sent once",
			payload:    []byte(strings.Repeat("a", 1000)),
			clientsCnt: uint(5),
			sentCnt:    uint(3),
		},
		{
			descrition: "single client and 4k payload sent once",
			payload:    []byte(strings.Repeat("a", 4096) + "\n"),
			clientsCnt: uint(1),
			sentCnt:    uint(1),
		},
	} {

		t.Run(tc.descrition, func(t *testing.T) {
			wg := sync.WaitGroup{}
			wg.Add(int(tc.clientsCnt))
			for i := 1; i <= int(tc.clientsCnt); i++ {
				go func(index int) {
					defer wg.Done()

					conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", echoPort))
					require.NoError(t, err)

					for sentCnt := 1; sentCnt <= int(tc.sentCnt); sentCnt++ {
						cntWrite, err := conn.Write([]byte(tc.payload))
						require.NoError(t, err)

						buff := make([]byte, 0, cntWrite)
						totalRead := 0
						tmp := make([]byte, cntWrite)
						for {
							cntRead, err := conn.Read(tmp)
							require.NoError(t, err)
							if cntRead > 0 {
								buff = append(buff, tmp[:cntRead]...)
								totalRead += cntRead
								if totalRead == cntWrite {
									break
								}
							} else {
								break
							}
						}

						require.NoError(t, err)
						require.Equal(t, cntWrite, totalRead)
						require.Equal(t, tc.payload, buff[0:totalRead])

						fmt.Printf("run %d completed (%d/%d) sent\n", index, sentCnt, tc.sentCnt)
					}

					fmt.Printf("run %d completed\n", index)
				}(i)
			}
			wg.Wait()
		})
	}
}
