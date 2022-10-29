package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
)

const (
	serverPort = 8886
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

	sessionId, err := uuid.NewUUID()
	if err != nil {
		fmt.Printf("session id generation failed - %v\n", err)
		return
	}

	bufSize := 9
	buffer := make([]byte, bufSize)
	session := make(map[int32]PriceRecord)
	//var payload []byte
	start := 0
	for {
		count, err := conn.Read(buffer[start:])
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%s\tRead error - %s\n", sessionId.String(), err)
			} else {
				fmt.Printf("%s\tClient closed connection\n", sessionId.String())
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
			fmt.Printf("%s\tfailed getting action: %v\n", sessionId.String(), err)
			continue
		}
		switch a := action.(type) {
		case PriceQuery:
			mean := calcMean(session, a)
			response := make([]byte, 4)
			MarshalInt32(mean, response)
			fmt.Printf("%s\treceived price query: %d %d \tcalculated mean: %d\n", sessionId.String(), a.MinTime, a.MaxTime, mean)
			writeCnt, err := conn.Write(response)
			if err != nil {
				fmt.Printf("%s\tfailed to write query response: %v", sessionId.String(), err)
				return
			}

			if writeCnt != len(response) {
				fmt.Printf("%s\tfailed to write all bytes of response: %v\n", sessionId.String(), err)
				return
			}

			continue
		case PriceRecord:
			session[a.Timestamp] = a
			fmt.Printf("%s\treceived price record: %d %d\n", sessionId.String(), a.Price, a.Timestamp)
			continue
		}
	}
}

func calcMean(session map[int32]PriceRecord, q PriceQuery) int32 {
	var mean int32 = 0
	var vals []int64
	for k, v := range session {
		if k >= q.MinTime && k <= q.MaxTime {
			vals = append(vals, int64(v.Price))
		}
	}

	vLen := int64(len(vals))
	if vLen > 0 {
		meanVal := int64(0)
		for _, v := range vals {
			meanVal += v
		}
		mean = int32(meanVal / vLen)
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
		timestamp := UnmarshalInt32(action[1:5])
		price := UnmarshalInt32(action[5:9])
		return PriceRecord{Timestamp: timestamp, Price: price}, nil
	case 'Q':
		// query
		minTime := UnmarshalInt32(action[1:5])
		maxTime := UnmarshalInt32(action[5:9])
		return PriceQuery{MinTime: minTime, MaxTime: maxTime}, nil
	}

	return nil, unrecognizedErr
}

type PriceRecord struct {
	Timestamp int32
	Price     int32
}

type PriceQuery struct {
	MinTime int32
	MaxTime int32
}

func MarshalInt32(v int32, b []byte) {
	buf := bytes.NewBuffer(make([]byte, 0, 4))
	binary.Write(buf, binary.BigEndian, v)
	copy(b, buf.Bytes())
}

func UnmarshalInt32(b []byte) int32 {
	var v int32
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.BigEndian, &v)
	return v
}
