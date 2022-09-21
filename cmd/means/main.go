package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"time"
)

const (
	serverPort = 8888
)

var (
	unrecognizedErr = errors.New("unrecognized action format")
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

	bufSize := 9
	buffer := make([]byte, bufSize)
	session := make(map[uint32]PriceRecord)
	//var payload []byte
	start := 0
	for {
		count, err := conn.Read(buffer[start:])
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error - %s\n", err)
			} else {
				fmt.Print("Client closed connection\n")
			}
			break
		}

		start += count
		if start < bufSize {
			continue
		} else {
			start = 0
		}

		actionBuff := buffer[:bufSize]
		action, err := getAction(actionBuff)
		if err != nil {
			fmt.Printf("failed getting action: %v", err)
			continue
		}
		switch a := action.(type) {
		case PriceQuery:
			mean := calcMean(session, a)
			response := make([]byte, 4)
			binary.BigEndian.PutUint32(response, mean)
			writeCnt, err := conn.Write(response)
			if err != nil {
				fmt.Printf("failed to write query response: %v", err)
				return
			}

			if writeCnt != len(response) {
				fmt.Printf("failed to write all bytes of response: %v", err)
				return
			}
			continue
		case PriceRecord:
			session[a.Timestamp] = a
			continue
		}
	}
}

func calcMean(session map[uint32]PriceRecord, q PriceQuery) uint32 {
	var mean uint32 = 0
	var vals []uint32
	for k, v := range session {
		if k >= q.MinTime && k <= q.MaxTime {
			vals = append(vals, v.Price)
		}
	}

	vLen := len(vals)

	sort.Slice(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})

	if vLen%2 == 0 {
		mean = (vals[vLen/2] + vals[(vLen/2)-1]) / 2
	} else {
		mean = vals[vLen/2]
	}

	return mean
}

func getAction(action []byte) (interface{}, error) {
	if len(action) != 9 {
		return nil, fmt.Errorf("malformed action, len: %d", len(action))
	}

	switch action[0] {
	case 'I':
		// insert
		timestamp := binary.BigEndian.Uint32(action[1:5])
		price := binary.BigEndian.Uint32(action[5:9])
		return PriceRecord{Timestamp: timestamp, Price: price}, nil
	case 'Q':
		// query
		minTime := binary.BigEndian.Uint32(action[1:5])
		maxTime := binary.BigEndian.Uint32(action[5:9])
		return PriceQuery{MinTime: minTime, MaxTime: maxTime}, nil
	}

	return nil, unrecognizedErr
}

type PriceRecord struct {
	Timestamp uint32
	Price     uint32
}

type PriceQuery struct {
	MinTime uint32
	MaxTime uint32
}
