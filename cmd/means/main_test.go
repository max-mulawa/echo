package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	server = "localhost"
)

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func TestWritingPrices(t *testing.T) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", server, serverPort))
	require.NoError(t, err)
	defer conn.Close()

	prices := []PriceRecord{
		{Timestamp: 12345, Price: 101},
		{Timestamp: 12346, Price: 102},
		{Timestamp: 12347, Price: 100},
		{Timestamp: 40960, Price: 5},
	}

	for _, p := range prices {
		for _, b := range MarshalPriceRecord(p) {
			conn.Write([]byte{b})
			//time.Sleep(time.Second * 3)
		}
	}

	conn.Write(MarshalPriceQuery(PriceQuery{MinTime: 12288, MaxTime: 16384}))
	meanResp := make([]byte, 4)
	conn.Read(meanResp)

	mean := binary.BigEndian.Uint32(meanResp)

	require.Equal(t, uint32(101), mean)

}

func MarshalPriceRecord(r PriceRecord) []byte {
	record := make([]byte, 9)
	record[0] = 'I'

	binary.BigEndian.PutUint32(record[1:5], r.Timestamp)
	binary.BigEndian.PutUint32(record[5:9], r.Price)

	return record
}

func MarshalPriceQuery(q PriceQuery) []byte {
	record := make([]byte, 9)
	record[0] = 'Q'

	binary.BigEndian.PutUint32(record[1:5], q.MinTime)
	binary.BigEndian.PutUint32(record[5:9], q.MaxTime)

	return record
}
