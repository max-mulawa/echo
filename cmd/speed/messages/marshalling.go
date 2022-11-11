package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"time"
)

// var msgTypes map[MsgType]reflect.Type = map[MsgType]reflect.Type{
// 	// ops.ErrorMsgType:            reflect.TypeOf(ops.ServerError{}),
// 	// ops.HeartbeatRequestMsgType: reflect.TypeOf(ops.HeartbeatRequest{}),
// 	// ops.HeartbeatMsgType:        reflect.TypeOf(ops.HearbeatSignal{}),
// }

type Decoder struct {
	msgTypes map[MsgType]reflect.Type
}

func NewDecoder() *Decoder {
	return &Decoder{msgTypes: make(map[MsgType]reflect.Type)}
}

func (d *Decoder) Unmarshall(payload []byte) (interface{}, int, error) {
	if len(payload) == 0 {
		return nil, 0, fmt.Errorf("empty payload")
	}

	msgType := MsgType(payload[0])
	var (
		t  reflect.Type
		ok bool
	)
	if t, ok = d.msgTypes[msgType]; !ok {
		return nil, 0, fmt.Errorf("message type (%d) not registered", msgType)
	}

	msg := reflect.Indirect(reflect.New(t))
	offset := 1

	if t.Kind() == reflect.Struct {
		for i := 0; i < msg.NumField(); i++ {
			switch msg.Field(i).Kind() {
			case reflect.Uint8:
				msg.Field(i).SetUint(uint64(payload[offset]))
				offset++
			case reflect.Uint16:
				var v uint16
				fromBigEndian(&v, payload[offset:offset+2])
				msg.Field(i).SetUint(uint64(v))
				offset += 2
			case reflect.Uint32:
				var v uint32
				fromBigEndian(&v, payload[offset:offset+4])
				msg.Field(i).SetUint(uint64(v))
				offset += 4
			case reflect.String:
				strLen := uint8(payload[offset])
				offset++
				bytesCnt := int(strLen)
				msg.Field(i).SetString(string(payload[offset:(offset + bytesCnt)]))
				offset += bytesCnt
			case reflect.Struct:
				switch msg.Field(i).Type() {
				case reflect.TypeOf(time.Time{}):
					var v uint32
					fromBigEndian(&v, payload[offset:offset+4])
					msg.Field(i).Set(reflect.ValueOf(time.Unix(int64(v), 0)))
					offset += 4
				}
				// time.Time
			case reflect.Slice:
				switch msg.Field(i).Type() {
				case reflect.TypeOf([]uint16{}):
					arrLen := uint8(payload[offset])
					offset++
					var v = make([]uint16, arrLen)
					bytesCnt := int(arrLen * 2)
					fromBigEndian(v, payload[offset:(offset+bytesCnt)])
					msg.Field(i).Set(reflect.ValueOf(v))
					offset += bytesCnt
				}
			}
		}
	}

	return msg.Interface(), offset, nil
}

func (d *Decoder) Marshal(val interface{}) ([]byte, error) {
	currentType := reflect.TypeOf(val)

	var msgType MsgType
	var mt reflect.Type
	for msgType, mt = range d.msgTypes {
		if currentType.ConvertibleTo(mt) {
			break
		}
	}

	if msgType == 0 {
		return nil, fmt.Errorf("message type not registered")
	}

	payload := make([]byte, 0)
	payload = append(payload, byte(msgType))

	v := reflect.ValueOf(val)

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			switch v.Field(i).Kind() {
			case reflect.Uint8:
				payload = append(payload, byte(v.Field(i).Uint()))
			case reflect.Uint16:
				payload = append(payload, toBigEndian(uint16(v.Field(i).Uint()))...)
			case reflect.Uint32:
				payload = append(payload, toBigEndian(uint32(v.Field(i).Uint()))...)
			case reflect.String:
				val := v.Field(i).String()
				strLen := uint8(len(val))
				payload = append(payload, byte(strLen))
				payload = append(payload, toBigEndian([]byte(val))...)
			case reflect.Struct:
				switch v.Field(i).Type() {
				case reflect.TypeOf(time.Time{}):
					var val uint32 = uint32(v.Field(i).Interface().(time.Time).Unix())
					payload = append(payload, toBigEndian(val)...)
				}
			case reflect.Slice:
				switch v.Field(i).Type() {
				case reflect.TypeOf([]uint16{}):
					val := v.Field(i).Interface().([]uint16)
					arrLen := uint8(len(val))
					payload = append(payload, byte(arrLen))
					payload = append(payload, toBigEndian(val)...)
				}
			}

		}

	}

	return payload, nil
}

func (d *Decoder) RegisterMsg(m MsgType, t reflect.Type) {
	d.msgTypes[m] = t
}

func fromBigEndian(v interface{}, b []byte) {
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.BigEndian, v)
}

func toBigEndian(v interface{}) (b []byte) {
	buf := bytes.NewBuffer(make([]byte, 0)) //?
	binary.Write(buf, binary.BigEndian, v)
	return buf.Bytes()
}
