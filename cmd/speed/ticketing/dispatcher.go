package ticketing

import (
	"fmt"
	"io"
	"max-mulawa/echo/cmd/speed/messages"
	"sync"

	"golang.org/x/exp/slices"
)

type IAmDispatcherMsg struct {
	Roads []uint16
}

var (
	IAmDispatcherMsgType messages.MsgType = 129 //0x81
)

type Dispatcher struct {
	roads   []uint16
	conn    io.Writer
	decoder *messages.Decoder
	Tickets chan TicketMsg
}

func NewDispatcher(m IAmDispatcherMsg, conn io.Writer, decoder *messages.Decoder) *Dispatcher {
	return &Dispatcher{
		roads:   m.Roads,
		conn:    conn,
		decoder: decoder,
		Tickets: make(chan TicketMsg),
	}
}

func (d *Dispatcher) Dispatch(t TicketMsg) error {
	payload, err := d.decoder.Marshal(t)
	if err != nil {
		return fmt.Errorf("dispatch marshalling failed: %w", err)
	}
	_, err = d.conn.Write(payload)
	if err != nil {
		return fmt.Errorf("dispatch sending payload failed: %w", err)
	}
	d.Tickets <- t
	return nil
}

type RoadDispatchers struct {
	lock        sync.Mutex
	dispatchers map[uint16][]*Dispatcher
	ticketQueue map[uint16][]TicketMsg
	ledger      *TicketLedger
}

func NewRoadDispatchers() *RoadDispatchers {
	return &RoadDispatchers{
		lock:        sync.Mutex{},
		dispatchers: make(map[uint16][]*Dispatcher),
		ticketQueue: map[uint16][]TicketMsg{},
		ledger:      &TicketLedger{},
	}
}

func (rd *RoadDispatchers) Register(d *Dispatcher) {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	for _, road := range d.roads {
		rd.dispatchers[road] = append(rd.dispatchers[road], d)
		if len(rd.ticketQueue[road]) > 0 {
			fmt.Printf("dispatching from the road (%d) queue\n", road)
			for _, t := range rd.ticketQueue[road] {
				rd.dispatchInternal(d, t)
			}
			rd.ticketQueue[road] = nil
		}
	}
}

func (rd *RoadDispatchers) Unregister(d *Dispatcher) {
	rd.lock.Lock()
	defer rd.lock.Unlock()
	for _, road := range d.roads {
		roadDisp := rd.dispatchers[road]
		for i, disp := range roadDisp {
			if d == disp {
				slices.Delete(roadDisp, i, i)
				break
			}
		}
	}
}

func (rd *RoadDispatchers) Dispatch(t TicketMsg) {
	rd.lock.Lock()
	defer rd.lock.Unlock()

	roadDispatchers, hasDispatchers := rd.dispatchers[t.Road]
	if hasDispatchers && len(roadDispatchers) > 0 {
		first := roadDispatchers[0]
		rd.dispatchInternal(first, t)
	} else {
		// queue when dispatcher handling this road registers
		// If the server generates a ticket for a road that has no connected dispatcher, it must store the ticket and deliver it once a dispatcher for that road is available.
		fmt.Println("dispaching ticket was postponed", t)
		rd.ticketQueue[t.Road] = append(rd.ticketQueue[t.Road], t)
	}
}

func (rd *RoadDispatchers) dispatchInternal(d *Dispatcher, t TicketMsg) {
	road := t.Road
	if rd.WasAddedToLedger(t) {
		err := d.Dispatch(t)
		if err != nil {
			fmt.Printf("failed to dispatch (road: %d) ticket (%v): %v\n", road, t, err)
		} else {
			fmt.Printf("dispatch ticket (road: %d) ticket (%v)\n", road, t)
		}
	} else {
		fmt.Printf("cannot dispatch ticket as car already fined on that day (road: %d) from the queue (%v)\n", road, t)
	}
}

func (rd *RoadDispatchers) WasAddedToLedger(t TicketMsg) bool {
	return rd.ledger.Add(t)
}
