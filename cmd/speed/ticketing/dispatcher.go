package ticketing

import (
	"max-mulawa/echo/cmd/speed/messages"
	"sync"
)

type IAmDispatcherMsg struct {
	Roads []uint16
}

var (
	IAmDispatcherMsgType messages.MsgType = 129 //0x81
)

type Dispatcher struct {
	roads []uint16
}

type RoadDispatchers struct {
	lock        sync.Mutex
	dispatchers map[uint16][]*Dispatcher
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
}
