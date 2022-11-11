package tracking

import (
	"max-mulawa/echo/cmd/speed/messages"
	"time"
)

type Camera struct {
	road  uint16 // road number
	mile  uint16 // relative position on the road
	limit uint16 // in miles per hour
}

type IAmCameraMsg struct {
	Camera
}

var IAmCameraMsgType messages.MsgType = 128 // 0x80

type Plate string

type Measurement struct {
	plate     Plate
	timestamp time.Time
}

var MeasurementMsgType messages.MsgType = 32 //0x20
