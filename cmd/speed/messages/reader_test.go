package messages_test

import (
	"bufio"
	"bytes"
	"max-mulawa/echo/cmd/speed/messages"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMessages(t *testing.T) {
	type stringTest struct {
		Msg string
	}

	buf := bufio.NewReader(bytes.NewBuffer([]byte{
		0x10, 0x03, 0x62, 0x61, 0x64,
		0x10, 0x03, 0x62, 0x61, 0x64,
		0x10, 0x03, 0x62, 0x61, 0x64,
	}))
	decoder := messages.NewDecoder()
	decoder.RegisterMsg(messages.MsgType(16), reflect.TypeOf(stringTest{}))

	r := messages.NewReader(buf, decoder)
	msgs := r.GetMessages()
	msgCounter := 0
	for m := range msgs {
		switch message := m.(type) {
		case stringTest:
			require.Equal(t, "bad", message.Msg)
			msgCounter++
		case error:
			require.Equal(t, messages.ErrClientClosed, message)
		default:
			require.Fail(t, "unsupported message type")
		}
	}

	require.Equal(t, 3, msgCounter)
}
