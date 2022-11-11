package ops

import "max-mulawa/echo/cmd/speed/messages"

var HeartbeatRequestMsgType messages.MsgType = 64 // 0x40

type HeartbeatRequest struct {
}

var HeartbeatMsgType messages.MsgType = 65 // 0x41

type HearbeatSignal struct {
}
