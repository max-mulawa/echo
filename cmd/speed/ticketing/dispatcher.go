package ticketing

import "max-mulawa/echo/cmd/speed/messages"

type RoadNum uint16

type IAmDispatcherMsg struct {
	roads []RoadNum
}

var ErrorMsgType messages.MsgType = 129 //0x81

type Dispatcher struct {
	roads []RoadNum
}

type RoadDispatchers struct {
	dispatchers map[RoadNum][]*Dispatcher
}

func (rd *RoadDispatchers) Register(d *Dispatcher) {
	for _, road := range d.roads {
		rd.dispatchers[road] = append(rd.dispatchers[road], d)
	}
}

func (rd *RoadDispatchers) Unregister(d *Dispatcher) {

}
