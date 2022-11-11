package messages_test

import (
	"max-mulawa/echo/cmd/speed/messages"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnmarshallingString(t *testing.T) {
	type stringTest struct {
		Msg string
	}

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(messages.MsgType(16), reflect.TypeOf(stringTest{}))

	payload := []byte{0x10, 0x03, 0x62, 0x61, 0x64}
	v, cnt, err := decoder.Unmarshall(payload)
	require.NoError(t, err)
	res := v.(stringTest)
	require.Equal(t, "bad", res.Msg)
	require.Equal(t, 5, cnt)
}

func TestUnmarshallingSignedInts(t *testing.T) {
	type uintTest struct {
		Interval uint32
		Road     uint16
		Roads    []uint16
	}

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(messages.MsgType(16), reflect.TypeOf(uintTest{}))

	payload := []byte{
		0x10,                   // message type
		0x00, 0x00, 0x00, 0x0a, // Internval uint32(10),
		0x00, 0x42, // Road: uint16(66),
		0x02,       // slice uint16[2]
		0x00, 0x42, //[0]=>66,
		0x01, 0x70, //[1]=>368,
	}
	v, cnt, err := decoder.Unmarshall(payload)
	require.NoError(t, err)
	res := v.(uintTest)
	require.Equal(t, uint32(10), res.Interval)
	require.Equal(t, uint16(66), res.Road)

	require.Equal(t, uint16(66), res.Roads[0])
	require.Equal(t, uint16(368), res.Roads[1])

	require.Equal(t, 12, cnt)
}

func TestUnmarshallingUnixTimestamp(t *testing.T) {
	type timestampTest struct {
		Timestamp time.Time
	}

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(messages.MsgType(16), reflect.TypeOf(timestampTest{}))

	payload := []byte{
		0x10,                   // message type
		0x00, 0x0f, 0x42, 0x40, // Timestamp (epoch + 1 000 000)
	}
	v, cnt, err := decoder.Unmarshall(payload)
	require.NoError(t, err)
	res := v.(timestampTest)
	require.Equal(t, time.Date(1970, 01, 12, 14, 46, 40, 0, time.Local), res.Timestamp)
	require.Equal(t, 5, cnt)
}

func TestUnmarshallingComplex(t *testing.T) {
	type complexTest struct {
		Plate     string
		Road      uint16
		Timestamp time.Time
	}

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(messages.MsgType(16), reflect.TypeOf(complexTest{}))

	payload := []byte{
		0x10,                         // message type
		0x04, 0x55, 0x4e, 0x31, 0x58, // plate: "UN1X"
		0x00, 0x42, // road: 66,
		0x00, 0x0f, 0x42, 0x40, // Timestamp (epoch + 1 000 000)
	}
	v, cnt, err := decoder.Unmarshall(payload)
	require.NoError(t, err)
	res := v.(complexTest)
	require.Equal(t, "UN1X", res.Plate)
	require.Equal(t, uint16(66), res.Road)
	require.Equal(t, time.Date(1970, 01, 12, 14, 46, 40, 0, time.Local), res.Timestamp)
	require.Equal(t, 12, cnt)
}

func TestMarshallingComplex(t *testing.T) {
	type complexTest struct {
		Plate     string
		Interval  uint32
		Road      uint16
		Timestamp time.Time
	}

	decoder := messages.NewDecoder()
	decoder.RegisterMsg(messages.MsgType(16), reflect.TypeOf(complexTest{}))

	payload := []byte{
		0x10,                         // message type
		0x04, 0x55, 0x4e, 0x31, 0x58, // plate: "UN1X"
		0x00, 0x00, 0x00, 0x0a, // Internval uint32(10),
		0x00, 0x42, // road: 66,
		0x00, 0x0f, 0x42, 0x40, // Timestamp (epoch + 1 000 000)
	}
	v, cnt, err := decoder.Unmarshall(payload)
	require.NoError(t, err)
	res := v.(complexTest)
	require.Equal(t, 16, cnt)

	resPayload, err := decoder.Marshal(res)
	require.NoError(t, err)
	require.Equal(t, payload, resPayload)
}
