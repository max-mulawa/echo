package messages

import (
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

		for {
			buffer := make([]byte, bufSize)
			start := 0
			count, err := r.conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					messages <- fmt.Errorf("read error: %w", err)
				} else {
					messages <- ErrClientClosed
				}
				break
			}

			for start < count {
				payload := buffer[start:count]
				msg, cntBytes, err := r.decoder.Unmarshall(payload)
				if err != nil {
					messages <- fmt.Errorf("failure during unmarshalling of message: %w", err)
					return
				}
				fmt.Printf("message size: %d\n", cntBytes)
				start += cntBytes
				messages <- msg
			}
		}
	}()
	return messages
}
