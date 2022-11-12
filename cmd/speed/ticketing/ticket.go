package ticketing

import (
	"max-mulawa/echo/cmd/speed/messages"
	"time"
)

type Ticket struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 time.Time
	Mile2      uint16
	Timestamp2 time.Time
	Speed      uint16
}

var TicketMsgType messages.MsgType = 33 //0x21
