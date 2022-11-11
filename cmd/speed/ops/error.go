package ops

import "max-mulawa/echo/cmd/speed/messages"

type ServerError struct {
	msg string
}

var ErrorMsgType messages.MsgType = 16 //0x10
