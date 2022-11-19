package ticketing_test

import (
	"bufio"
	"max-mulawa/echo/cmd/speed/messages"
	"max-mulawa/echo/cmd/speed/ticketing"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDispatcher(t *testing.T) {
	decoder := messages.NewDecoder()
	decoder.RegisterMsg(ticketing.TicketMsgType, reflect.TypeOf(ticketing.TicketMsg{}))
	dispatchers := ticketing.NewRoadDispatchers()
	road1 := uint16(1)
	road2 := uint16(2)
	roads := []uint16{road1, road2}
	m := ticketing.IAmDispatcherMsg{Roads: roads}

	builder := &strings.Builder{}
	conn := bufio.NewWriter(builder)
	dispatcher := ticketing.NewDispatcher(m, conn, decoder)
	dispatchers.Register(dispatcher)

	go dispatchers.Dispatch(ticketing.TicketMsg{
		Plate:      "abc",
		Road:       1,
		Timestamp1: time.Now(),
		Timestamp2: time.Now().Add(time.Minute * 30),
	})

	go dispatchers.Dispatch(ticketing.TicketMsg{
		Plate:      "abc",
		Road:       2,
		Timestamp1: time.Now().AddDate(0, 0, 1),
		Timestamp2: time.Now().AddDate(0, 0, 1).Add(time.Hour),
	})

	ticket := <-dispatcher.Tickets
	require.Subset(t, roads, []uint16{ticket.Road})

	ticket = <-dispatcher.Tickets
	require.Subset(t, roads, []uint16{ticket.Road})
}
