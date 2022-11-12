package tracking

import (
	"max-mulawa/echo/cmd/speed/messages"
	"time"
)

type IAmCameraMsg struct {
	Road  uint16 // road number
	Mile  uint16 // relative position on the road
	Limit uint16 // in miles per hour
}

var IAmCameraMsgType messages.MsgType = 128 // 0x80

type MeasurementTimeMsg struct {
	Plate     string
	Timestamp time.Time
}

var MeasurementTimeMsgType messages.MsgType = 32 //0x20

type Measurement struct {
	Device IAmCameraMsg
	Time   MeasurementTimeMsg
}
