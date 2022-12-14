package messages

import (
	"encoding/hex"
	"fmt"
	"io"
)

var (
	ErrClientClosed = fmt.Errorf("client closed connection")
)

type Reader struct {
	conn    io.Reader
	decoder *Decoder
}

func NewReader(conn io.Reader, decoder *Decoder) *Reader {
	return &Reader{
		conn:    conn,
		decoder: decoder,
	}
}

func (r *Reader) GetMessages() <-chan interface{} {
	messages := make(chan interface{})
	go func() {
		defer close(messages)
		bufSize := 1024
		incomplete := []byte{}

		for {
			buffer := make([]byte, bufSize)
			start := 0
			readCnt, err := r.conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					messages <- fmt.Errorf("read error: %w", err)
				} else {
					messages <- ErrClientClosed
				}
				break
			}

			payloadBuf := append(incomplete, buffer[start:readCnt]...)
			payloadBufCnt := len(payloadBuf)
			incomplete = []byte{}

			for start < payloadBufCnt {
				payload := payloadBuf[start:]
				msg, cntBytes, err := r.decoder.Unmarshall(payload)
				if err == ErrIncompletePayload {
					incomplete = payload
					break
				} else if err != nil {
					messages <- fmt.Errorf("failure during unmarshalling of message: %w", err)
					return
				}
				fmt.Printf("message size: %d, payload: %+v\n", cntBytes, hex.EncodeToString(payload[:cntBytes]))
				start += cntBytes
				messages <- msg
			}
		}
	}()
	return messages
}
