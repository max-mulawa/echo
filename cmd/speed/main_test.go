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
	camera := Connect(t)
	defer camera.Close()

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))

	payload, err := decoder.Marshal(tracking.IAmCameraMsg{
		Road:  124,
		Mile:  8,
		Limit: 60,
	})
	require.NoError(t, err)

	Write(t, camera, payload)
}

func TestRegisterDispatcher(t *testing.T) {
	conn := Connect(t)
	defer conn.Close()

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(ticketing.IAmDispatcherMsgType, reflect.TypeOf(ticketing.IAmDispatcherMsg{}))

	payload, err := decoder.Marshal(ticketing.IAmDispatcherMsg{
		Roads: []uint16{8, 125, 4},
	})
	require.NoError(t, err)
	Write(t, conn, payload)
}

func TestRegisterCameraTwice(t *testing.T) {
	camera := Connect(t)
	defer camera.Close()

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))
	decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))

	payload, err := decoder.Marshal(tracking.IAmCameraMsg{
		Road:  126,
		Mile:  8,
		Limit: 60,
	})
	require.NoError(t, err)

	Write(t, camera, payload)
	Write(t, camera, payload)

	reader := messages.NewReader(camera, decoder)
	msgs := reader.GetMessages()
	srvErr := (<-msgs).(ops.ServerError)
	require.Contains(t, srvErr.Msg, "camera")
	require.Contains(t, srvErr.Msg, "incorrect order")

	buf := make([]byte, 1024)
	_, err = camera.Read(buf)
	require.Equal(t, io.EOF, err)
}

func TestRegisterCameraSendMeasurementsAndWaitForTicket(t *testing.T) {
	carPlate := "X1334"
	road := uint16(0)
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

func TestMeasurementWithInvalidNumberPlate(t *testing.T) {
	invalidCarPlate := "X1334abc"
	road := uint16(122)
	mile1 := uint16(9)
	speedLimit := uint16(70)

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))
	decoder.RegisterMsg(tracking.MeasurementTimeMsgType, reflect.TypeOf(tracking.MeasurementTimeMsg{}))
	decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))

	// register cameras
	registerCamera1, err := decoder.Marshal(tracking.IAmCameraMsg{
		Road:  road,
		Mile:  mile1,
		Limit: speedLimit,
	})
	require.NoError(t, err)

	cam1Measurement, err := decoder.Marshal(tracking.MeasurementTimeMsg{
		Plate:     invalidCarPlate,
		Timestamp: time.Now().Add(-10 * time.Minute),
	})
	require.NoError(t, err)

	camera := Connect(t)
	defer camera.Close()

	Write(t, camera, append(registerCamera1, cam1Measurement...))

	reader := messages.NewReader(camera, decoder)
	msgs := reader.GetMessages()
	srvErr := (<-msgs).(ops.ServerError)
	require.Contains(t, srvErr.Msg, invalidCarPlate)
}

func TestUnknownMessageType(t *testing.T) {
	type stringTest struct {
		Msg string
	}

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(messages.MsgType(5), reflect.TypeOf(stringTest{}))
	decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))

	unknownMsg, err := decoder.Marshal(stringTest{
		Msg: "i'm unknown type",
	})
	require.NoError(t, err)

	require.NoError(t, err)

	client := Connect(t)
	defer client.Close()

	Write(t, client, unknownMsg)

	reader := messages.NewReader(client, decoder)
	msgs := reader.GetMessages()
	srvErr := (<-msgs).(ops.ServerError)
	require.Contains(t, srvErr.Msg, "unknown")
}

func TestUnorderedMessages(t *testing.T) {
	decoder := messages.NewDecoder()
	decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))
	decoder.RegisterMsg(tracking.MeasurementTimeMsgType, reflect.TypeOf(tracking.MeasurementTimeMsg{}))

	measurementMsg, err := decoder.Marshal(tracking.MeasurementTimeMsg{
		Plate:     "ABC123",
		Timestamp: time.Now().Add(-10 * time.Minute),
	})
	require.NoError(t, err)

	client := Connect(t)
	defer client.Close()

	Write(t, client, measurementMsg)

	reader := messages.NewReader(client, decoder)
	msgs := reader.GetMessages()
	srvErr := (<-msgs).(ops.ServerError)
	require.Contains(t, srvErr.Msg, "message out of scope")
}

func TestHearbeatMessages(t *testing.T) {
	decoder := messages.NewDecoder()
	decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))
	decoder.RegisterMsg(ops.HeartbeatRequestMsgType, reflect.TypeOf(ops.HeartbeatRequest{}))
	decoder.RegisterMsg(ops.HeartbeatMsgType, reflect.TypeOf(ops.HearbeatSignal{}))

	heartbeatReq, err := decoder.Marshal(ops.HeartbeatRequest{
		Interval: 3,
	})
	require.NoError(t, err)

	client := Connect(t)
	defer client.Close()

	start := time.Now()
	Write(t, client, heartbeatReq)
	reader := messages.NewReader(client, decoder)
	msgs := reader.GetMessages()
	hearbeatSignal := (<-msgs).(ops.HearbeatSignal)
	diff := time.Since(start)
	require.Equal(t, ops.HearbeatSignal{}, hearbeatSignal)
	require.GreaterOrEqual(t, diff, time.Duration(0.3*time.Hour.Seconds()))
}

func TestSampleSession(t *testing.T) {
	camera := Connect(t)
	defer camera.Close()
	camera2 := Connect(t)
	defer camera2.Close()
	dispatcher := Connect(t)
	defer dispatcher.Close()

	// 80 00 7b 00 08 00 3c	<- IAmCamera{road: 123, mile: 8, limit: 60}
	registerCamera1 := []byte{0x80, 0x00, 0x7b, 0x00, 0x08, 0x00, 0x3c}
	// 20 04 55 4e 31 58 00 00 00 00 Plate{plate: "UN1X", timestamp: 0}
	cam1Measurement := []byte{0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x00, 0x00}

	// 80 00 7b 00 09 00 3c <-- IAmCamera{road: 123, mile: 9, limit: 60}
	registerCamera2 := []byte{0x80, 0x00, 0x7b, 0x00, 0x09, 0x00, 0x3c}
	// 20 04 55 4e 31 58 00 00 00 2d <-- Plate{plate: "UN1X", timestamp: 45}
	cam2Measurement := []byte{0x20, 0x04, 0x55, 0x4e, 0x31, 0x58, 0x00, 0x00, 0x00, 0x2d}

	// 81 01 00 7b <-- IAmDispatcher{roads: [123]}
	registerDispatcher := []byte{0x81, 0x01, 0x00, 0x7b}
	// 21 04 55 4e 31 58 00 7b 00 08 00 00 00 00 00 09 00 00 00 2d 1f 40 <-- Ticket{plate: "UN1X", road: 123, mile1: 8, timestamp1: 0, mile2: 9, timestamp2: 45, speed: 8000}
	// := []byte{}

	Write(t, camera, append(registerCamera1, cam1Measurement...))
	Write(t, camera2, append(registerCamera2, cam2Measurement...))
	Write(t, dispatcher, registerDispatcher)

	// time.Sleep(time.Second)

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(ticketing.TicketMsgType, reflect.TypeOf(ticketing.TicketMsg{}))
	reader := messages.NewReader(dispatcher, decoder)
	msgs := reader.GetMessages()
	ticket := (<-msgs).(ticketing.TicketMsg)
	require.Equal(t, uint16(123), ticket.Road)
	require.Equal(t, "UN1X", ticket.Plate)
	require.Equal(t, uint16(8), ticket.Mile1)
	require.Equal(t, time.Unix(0, 0), ticket.Timestamp1)
	require.Equal(t, uint16(9), ticket.Mile2)
	require.Equal(t, time.Unix(45, 0), ticket.Timestamp2)
	require.Equal(t, uint16(8000), ticket.Speed)
}

// TODO: dispatcher responsible for multiple roads
// TODO: not present dispatcher to ticket on the road
// TODO: disconnected dispatcher
// TODO: Only 1 ticket per car per day
// TODO: Observation spanning 2 days covers tickets in both days.

func Connect(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "localhost", serverPort))
	require.NoError(t, err)
	return conn
}

func Write(t *testing.T, conn net.Conn, payload []byte) {
	cntWrite, err := conn.Write(payload)
	require.NoError(t, err)
	require.Equal(t, len(payload), cntWrite)
}
