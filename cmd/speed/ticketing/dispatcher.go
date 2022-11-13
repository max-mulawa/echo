package ticketing

import (
	"fmt"
	"max-mulawa/echo/cmd/speed/messages"
	"net"
	"sync"
)

type IAmDispatcherMsg struct {
	Roads []uint16
}

var (
	IAmDispatcherMsgType messages.MsgType = 129 //0x81
)

type Dispatcher struct {
	roads   []uint16
	conn    net.Conn
	decoder *messages.Decoder
	Tickets chan TicketMsg
}

func NewDispatcher(m IAmDispatcherMsg, conn net.Conn, decoder *messages.Decoder) *Dispatcher {
	return &Dispatcher{
		roads:   m.Roads,
		conn:    conn,
		decoder: decoder,
		Tickets: make(chan TicketMsg),
	}
}

func (d *Dispatcher) Dispatch(t TicketMsg) {
	payload, err := d.decoder.Marshal(t)
	if err != nil {
		fmt.Printf("dispatch marshalling failed: %v\n", err)
	}
	_, err = d.conn.Write(payload)
	if err != nil {
		fmt.Printf("dispatch sending payload failed: %v\n", err)
	}
	d.Tickets <- t
}

type RoadDispatchers struct {
	lock        sync.Mutex
	dispatchers map[uint16][]*Dispatcher
}

func NewRoadDispatchers() *RoadDispatchers {
	return &RoadDispatchers{
		lock:        sync.Mutex{},
		dispatchers: make(map[uint16][]*Dispatcher),
	}
}

func (rd *RoadDispatchers) Register(d *Dispatcher) {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	for _, road := range d.roads {
		rd.dispatchers[road] = append(rd.dispatchers[road], d)
	}
}

func (rd *RoadDispatchers) Unregister(d *Dispatcher) {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	// remove dispatecher from Roads
	// for _, _ := range d.roads {
	// 	// TODO :slices.Delete(rd.dispatchers[road], d)
	// }
}

func (rd *RoadDispatchers) Dispatch(t TicketMsg) {
	// check in the ledger if there is already a ticket for this day, for this car
	// floor(timestamp / 86400) = day

	roadDispatchers := rd.dispatchers[t.Road]
	if roadDispatchers != nil {
		first := roadDispatchers[0]
		first.Dispatch(t)
	} else {
		// queue when dispatcher handling this road registers
		// If the server generates a ticket for a road that has no connected dispatcher, it must store the ticket and deliver it once a dispatcher for that road is available.
		fmt.Println("dispaching ticket was postponed", t)
	}
}
