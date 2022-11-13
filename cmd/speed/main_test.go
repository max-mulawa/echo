package main

import (
	"fmt"
	"io"
	"max-mulawa/echo/cmd/speed/messages"
	"max-mulawa/echo/cmd/speed/ops"
	"max-mulawa/echo/cmd/speed/ticketing"
	"max-mulawa/echo/cmd/speed/tracking"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	go startServer()
	m.Run()
}

func TestRegisterCamera(t *testing.T) {
	conn := Connect(t)
	defer conn.Close()

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))

	payload, err := decoder.Marshal(tracking.IAmCameraMsg{
		Road:  123,
		Mile:  8,
		Limit: 60,
	})
	require.NoError(t, err)

	Write(t, conn, payload)
}

func TestRegisterDispatcher(t *testing.T) {
	conn := Connect(t)
	defer conn.Close()

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(ticketing.IAmDispatcherMsgType, reflect.TypeOf(ticketing.IAmDispatcherMsg{}))

	payload, err := decoder.Marshal(ticketing.IAmDispatcherMsg{
		Roads: []uint16{8, 123, 4},
	})
	require.NoError(t, err)
	Write(t, conn, payload)
}

func TestRegisterCameraTwice(t *testing.T) {
	conn := Connect(t)
	defer conn.Close()

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))
	decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))

	payload, err := decoder.Marshal(tracking.IAmCameraMsg{
		Road:  123,
		Mile:  8,
		Limit: 60,
	})
	require.NoError(t, err)

	Write(t, conn, payload)
	Write(t, conn, payload)

	buf := make([]byte, 1024)
	byteCnt, err := conn.Read(buf)
	require.NoError(t, err)
	msg, _, err := decoder.Unmarshall(buf[:byteCnt])
	require.NoError(t, err)
	switch emsg := msg.(type) {
	case ops.ServerError:
		require.Contains(t, emsg.Msg, "camera")
		require.Contains(t, emsg.Msg, "incorrect order")
	default:
		require.Fail(t, "error message should be returned")
	}
	_, err = conn.Read(buf)
	require.Equal(t, io.EOF, err)
}

func TestRegisterCameraSendMeasurementsAndWaitForTicket(t *testing.T) {
	carPlate := "X1334"
	road := uint16(123)
	mile1 := uint16(8)
	mile2 := uint16(30)
	speedLimit := uint16(60)

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))
	decoder.RegisterMsg(tracking.MeasurementTimeMsgType, reflect.TypeOf(tracking.MeasurementTimeMsg{}))
	decoder.RegisterMsg(ticketing.IAmDispatcherMsgType, reflect.TypeOf(ticketing.IAmDispatcherMsg{}))
	decoder.RegisterMsg(ticketing.TicketMsgType, reflect.TypeOf(ticketing.TicketMsg{}))

	// register cameras
	registerCamera1, err := decoder.Marshal(tracking.IAmCameraMsg{
		Road:  road,
		Mile:  mile1,
		Limit: speedLimit,
	})
	require.NoError(t, err)

	registerCamera2, err := decoder.Marshal(tracking.IAmCameraMsg{
		Road:  road,
		Mile:  mile2,
		Limit: speedLimit,
	})
	require.NoError(t, err)

	cam1Measurement, err := decoder.Marshal(tracking.MeasurementTimeMsg{
		Plate:     carPlate,
		Timestamp: time.Now().Add(-10 * time.Minute),
	})
	require.NoError(t, err)

	cam2Measurement, err := decoder.Marshal(tracking.MeasurementTimeMsg{
		Plate:     carPlate,
		Timestamp: time.Now(),
	})
	require.NoError(t, err)

	registerDispatcher, err := decoder.Marshal(ticketing.IAmDispatcherMsg{
		Roads: []uint16{8, road, 4},
	})
	require.NoError(t, err)

	camera := Connect(t)
	defer camera.Close()
	camera2 := Connect(t)
	defer camera2.Close()
	dispatcher := Connect(t)
	defer dispatcher.Close()

	Write(t, dispatcher, registerDispatcher)
	Write(t, camera2, append(registerCamera2, cam2Measurement...))
	Write(t, camera, append(registerCamera1, cam1Measurement...))

	reader := messages.NewReader(dispatcher, decoder)
	msgs := reader.GetMessages()
	ticket := (<-msgs).(ticketing.TicketMsg)
	require.Equal(t, road, ticket.Road)
	require.Equal(t, carPlate, ticket.Plate)
	require.Equal(t, mile1, ticket.Mile1)
	require.Equal(t, mile2, ticket.Mile2)
	require.Equal(t, uint16(6*22*100), ticket.Speed)
}

func Connect(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "localhost", serverPort))
	require.NoError(t, err)
	return conn
}

func Write(t *testing.T, conn net.Conn, payload []byte) {
	cntWrite, err := conn.Write(payload)
	require.NoError(t, err)
	require.Equal(t, len(payload), cntWrite)
	time.Sleep(10 * time.Millisecond)
}
