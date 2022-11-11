package ticketing

import (
	"max-mulawa/echo/cmd/speed/messages"
	"time"
)

type Ticket struct {
	plate      string
	road       RoadNum
	mile1      uint16
	timestamp1 time.Time
	mile2      uint16
	timestamp2 time.Time
	speed      uint16
}

var TicketMsgType messages.MsgType = 33 //0x21
