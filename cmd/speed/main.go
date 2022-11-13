package main

import (
	"fmt"
	"max-mulawa/echo/cmd/speed/messages"
	"max-mulawa/echo/cmd/speed/ops"
	"max-mulawa/echo/cmd/speed/ticketing"
	"max-mulawa/echo/cmd/speed/tracking"
	"max-mulawa/echo/cmd/speed/traffic"
	"net"
	"os"
	"reflect"
	"time"
)

const (
	serverPort = 8806
)

var (
	offenseSub      = make(chan traffic.Offense)
	offenceFeed     = traffic.NewOffenseFeed(offenseSub)
	dispatchers     = ticketing.NewRoadDispatchers()
	measurementsReg = traffic.NewMeasurementsRegistry(offenceFeed)
)

func main() {
	startServer()
}

func startServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		fmt.Printf("failed to listen on port %d port: %v\n", serverPort, err)
		os.Exit(1)
	}

	go listenForOffences(dispatchers, offenseSub)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		conn.SetDeadline(time.Now().Add(time.Second * 120 * 10))
		go handleConnection(conn)
	}
}

func listenForOffences(dispatchers *ticketing.RoadDispatchers, offenses <-chan traffic.Offense) {
	for o := range offenses {
		t := ticketing.TicketMsg{
			Plate:      o.Plate,
			Road:       o.Road,
			Mile1:      o.Mile1,
			Timestamp1: o.Timestamp1,
			Mile2:      o.Mile2,
			Timestamp2: o.Timestamp2,
			Speed:      o.Speed,
		}
		dispatchers.Dispatch(t)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		fmt.Print("Closing source connection on server\n")
		conn.Close()
	}()

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(tracking.IAmCameraMsgType, reflect.TypeOf(tracking.IAmCameraMsg{}))
	decoder.RegisterMsg(tracking.MeasurementTimeMsgType, reflect.TypeOf(tracking.MeasurementTimeMsg{}))

	decoder.RegisterMsg(ticketing.IAmDispatcherMsgType, reflect.TypeOf(ticketing.IAmDispatcherMsg{}))
	decoder.RegisterMsg(ticketing.TicketMsgType, reflect.TypeOf(ticketing.TicketMsg{}))

	decoder.RegisterMsg(ops.ErrorMsgType, reflect.TypeOf(ops.ServerError{}))

	reader := messages.NewReader(conn, decoder)
	var msgHanlder Handler
	for msg := range reader.GetMessages() {
		if msgHanlder != nil {
			err := msgHanlder.Handle(msg)
			if err != nil {
				writeServerError(conn, decoder, fmt.Sprintf("handler failed message: %v", err))
				return
			}
			continue
		}

		switch m := msg.(type) {
		case tracking.IAmCameraMsg:
			msgHanlder = NewCameraHanlder(m, measurementsReg, decoder)
		case ticketing.IAmDispatcherMsg:
			msgHanlder = NewDispatcherHandler(conn, m, dispatchers, decoder)
		case error:
			fmt.Println("error occured", m)
			return
		default:
			writeServerError(conn, decoder, fmt.Sprintf("message out of scope: %v", msg))
			return
		}
	}

}

func writeServerError(conn net.Conn, decoder *messages.Decoder, message string) {
	fmt.Println(message)
	errPayload, _ := decoder.Marshal(ops.ServerError{Msg: message})
	conn.Write(errPayload)
}

func (d *DispatcherHanlder) Handle(msg interface{}) error {
	for t := range d.dispatcher.Tickets {
		fmt.Println("ticket dispatched", t)
	}
	return nil
}

type Handler interface {
	Handle(msg interface{}) error
}

type CameraHanlder struct {
	registry *traffic.MeasurementsRegistry
	camera   *tracking.Camera
	decoder  *messages.Decoder
}

func (c *CameraHanlder) Handle(msg interface{}) error {
	cam := c.camera.Metadata
	switch message := msg.(type) {
	case tracking.MeasurementTimeMsg:
		c.registry.Register(tracking.Measurement{Device: cam, Time: message})
		/// TODO: heartbeat message
	case error:
		if msg == messages.ErrClientClosed {
			fmt.Printf("camera (r: %d, m: %d): %s\n", cam.Road, cam.Mile, msg)
			return nil
		}
		return fmt.Errorf("incorrect message send to camera: %w", message)
	default:
		return fmt.Errorf("incorrect order of messages send to camera: %v", message)
	}
	return nil
}

func NewCameraHanlder(m tracking.IAmCameraMsg, registry *traffic.MeasurementsRegistry, decoder *messages.Decoder) *CameraHanlder {
	return &CameraHanlder{
		registry: registry,
		camera:   tracking.NewCamera(m),
		decoder:  decoder,
	}
}

type DispatcherHanlder struct {
	dispatcher *ticketing.Dispatcher
}

func NewDispatcherHandler(conn net.Conn, m ticketing.IAmDispatcherMsg, dispatchers *ticketing.RoadDispatchers, decoder *messages.Decoder) *DispatcherHanlder {
	dispatcher := ticketing.NewDispatcher(m, conn, decoder)
	dispatchers.Register(dispatcher)

	//TODO: unregister dispatcher

	return &DispatcherHanlder{
		dispatcher: dispatcher,
	}
}

// Accept connections
// - from Camera (Register, receives Measurements, Unregister when error or disconnected or ....)
// - from Ticket Dispatcher (registers for multiple roads, send tickets )
