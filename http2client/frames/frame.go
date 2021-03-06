package frames

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Type byte
type Flag byte

const (
	DATA_TYPE          Type = 0x00
	HEADERS_TYPE       Type = 0x01
	PRIORITY_TYPE      Type = 0x02
	RST_STREAM_TYPE    Type = 0x03
	SETTINGS_TYPE      Type = 0x04
	PUSH_PROMISE_TYPE  Type = 0x05
	PING_TYPE          Type = 0x06
	GOAWAY_TYPE        Type = 0x07
	WINDOW_UPDATE_TYPE Type = 0x08
	// TODO: CONTINUATION_TYPE
)

type Frame interface {
	Encode(*EncodingContext) ([]byte, error)
	Type() Type
	GetStreamId() uint32
}

func FindDecoder(frameType Type) func(flags byte, streamId uint32, payload []byte, context *DecodingContext) (Frame, error) {
	switch frameType {
	case DATA_TYPE:
		return DecodeDataFrame
	case HEADERS_TYPE:
		return DecodeHeadersFrame
	case PRIORITY_TYPE:
		return DecodePriorityFrame
	case RST_STREAM_TYPE:
		return DecodeRstStreamFrame
	case SETTINGS_TYPE:
		return DecodeSettingsFrame
	case PUSH_PROMISE_TYPE:
		return DecodePushPromiseFrame
	case PING_TYPE:
		return DecodePingFrame
	case GOAWAY_TYPE:
		return DecodeGoAwayFrame
	case WINDOW_UPDATE_TYPE:
		return DecodeWindowUpdateFrame
	default:
		return nil
	}
}

func encodeHeader(frameType Type, streamId uint32, length uint32, flags []Flag) []byte {
	var result bytes.Buffer
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, length)
	result.Write(bytes[1:])
	bytes[0] = byte(frameType)
	result.Write(bytes[0:1])
	bytes[0] = 0
	for _, flag := range flags {
		flag.set(&bytes[0])
	}
	result.Write(bytes[0:1])
	binary.BigEndian.PutUint32(bytes, streamId)
	result.Write(bytes)
	return result.Bytes()
}

type FrameHeader struct {
	Length     uint32
	HeaderType Type
	Flags      byte
	StreamId   uint32
}

func DecodeHeader(data []byte) *FrameHeader {
	return &FrameHeader{
		Length:     uint32_ignoreFirstBit(data[0:3]),
		HeaderType: Type(data[3]),
		Flags:      data[4],
		StreamId:   uint32_ignoreFirstBit(data[5:9]),
	}
}

func (flag Flag) isSet(flagsByte byte) bool {
	return flagsByte&byte(flag) != 0
}

func (flag Flag) set(flagsByte *byte) {
	*flagsByte = *flagsByte | byte(flag)
}

func stripPadding(payload []byte) ([]byte, error) {
	padLength := int(payload[0])
	if len(payload) <= padLength {
		// TODO: trigger connection error.
		return nil, fmt.Errorf("Invalid HEADERS frame: padding >= payload.")
	}
	return payload[1 : len(payload)-padLength], nil
}

func (t Type) String() string {
	switch t {
	case DATA_TYPE:
		return "DATA"
	case HEADERS_TYPE:
		return "HEADERS"
	case PRIORITY_TYPE:
		return "PRIORITY"
	case RST_STREAM_TYPE:
		return "RST_STREAM"
	case SETTINGS_TYPE:
		return "SETTINGS"
	case PUSH_PROMISE_TYPE:
		return "PUSH_PROMISE"
	case PING_TYPE:
		return "PING"
	case GOAWAY_TYPE:
		return "GOAWAY"
	case WINDOW_UPDATE_TYPE:
		return "WINDOW_UPDATE"
	default:
		return fmt.Sprintf("'UNKNOWN TYPE 0x%02X'", byte(t))
	}
}
