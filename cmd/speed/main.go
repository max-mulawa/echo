package main

import (
	"fmt"
	"io"
	"max-mulawa/echo/cmd/speed/messages"
	"max-mulawa/echo/cmd/speed/ops"
	"max-mulawa/echo/cmd/speed/ticketing"
	"max-mulawa/echo/cmd/speed/tracking"
	"net"
	"os"
	"reflect"
	"time"
)

const (
	serverPort = 8806
)

func main() {
	startServer()
}

func startServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		fmt.Printf("failed to listen on port %d port: %v", serverPort, err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		conn.SetDeadline(time.Now().Add(time.Second * 120 * 10))
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Print("Closing source connection on server\n")
		conn.Close()
	}()

	bufSize := 1024

	for {
		buffer := make([]byte, bufSize)
		start := 0
		count, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error - %s\n", err)
			} else {
				fmt.Print("Client closed connection\n")
			}
			break
		}

		payload := buffer[start:count]

		decoder := messages.NewDecoder()
		decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))
		decoder.RegisterMsg(ticketing.IAmDispatcherMsgType, reflect.TypeOf(ticketing.IAmDispatcherMsg{}))
		decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))

		msg, cntBytes, err := decoder.Unmarshall(payload)
		if err != nil {
			fmt.Printf("failure during unmarshalling of message: %v", err)
			errPayload, _ := decoder.Marshal(ops.ServerError{Msg: "Unsupported message type"})
			conn.Write(errPayload)
			return
		}
		fmt.Printf("message size: %d", cntBytes)

		switch m := msg.(type) {
		case tracking.IAmCameraMsg:
			onCameraRegister(conn, m)
		case ticketing.IAmDispatcherMsg:
			onDispatcherRegister(conn, m)
		default:
			fmt.Printf("incorrect message order, message (type: %v) unexpected", payload[0])
			errPayload, _ := decoder.Marshal(ops.ServerError{Msg: "Incorrect order of messages"})
			conn.Write(errPayload)
			return
		}
	}
}

func onDispatcherRegister(conn net.Conn, m ticketing.IAmDispatcherMsg) {
	panic("unimplemented")
}

func onCameraRegister(conn net.Conn, m tracking.IAmCameraMsg) {
	panic("unimplemented")
}

// Accept connections
// - from Camera (Register, receives Measurements, Unregister when error or disconnected or ....)
// - from Ticket Dispatcher (registers for multiple roads, send tickets )
