package structure

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHandshake(t *testing.T) {
	Convey("Parse a handshake message from bytes", t, func() {
		msg := "\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		br := bytes.NewReader([]byte(msg))
		hs, err := ReadHandshake(br)
		So(hs, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	// "Reject a malformed handshake message"
	Convey("Parse a handshake message from bytes", t, func() {
		msg := "\x13\x41\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		br := bytes.NewReader([]byte(msg))
		hs, err := ReadHandshake(br)
		So(hs, ShouldBeNil)
		So(err, ShouldEqual, ErrNotBitTorrentProtocol)
	})

	Convey("Convert a handshake message to bytes", t, func() {
		msg := "\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		br := bytes.NewReader([]byte(msg))
		hs, _ := ReadHandshake(br)
		b := hs.Bytes()
		So(b, ShouldNotBeNil)
		So(b, ShouldResemble, []byte(msg))
	})
}

type StringMessageTest struct {
	Desc    string
	String  string
	Message Message
}

func TestMessages(t *testing.T) {

	bf := BitFieldFromHexString("\xff\xff\xff\x01")

	smTests := []StringMessageTest{
		{"KeepAlive", "\x00\x00\x00\x00",
			&BasicMessage{Type: MessageTypeKeepAlive, Length: 0}},
		{"Choke", "\x00\x00\x00\x01\x00",
			&BasicMessage{Type: MessageTypeChoke, Length: 1}},
		{"Unchoke", "\x00\x00\x00\x01\x01",
			&BasicMessage{Type: MessageTypeUnchoke, Length: 1}},
		{"Interested", "\x00\x00\x00\x01\x02",
			&BasicMessage{Type: MessageTypeInterested, Length: 1}},
		{"NotInterested", "\x00\x00\x00\x01\x03",
			&BasicMessage{Type: MessageTypeNotInterested, Length: 1}},
		{"Have", "\x00\x00\x00\x05\x04\x00\x00\x18\xa4",
			&HaveMessage{BasicMessage: BasicMessage{Type: MessageTypeHave, Length: 5, Payload: []byte("\x00\x00\x18\xa4")}, PieceIndex: 6308}},
		{"BitField", "\x00\x00\x00\x05\x05\xff\xff\xff\x01",
			&BitFieldMessage{BasicMessage: BasicMessage{Type: MessageTypeBitField, Length: 5, Payload: []byte("\xff\xff\xff\x01")}, BitField: bf}},
		{"Request", "\x00\x00\x00\x0d\x06\x00\x00\x0b\xb0\x00\x02\x40\x00\x00\x00\x40\x00",
			&RequestMessage{BasicMessage: BasicMessage{Type: MessageTypeRequest, Length: 13, Payload: []byte("\x00\x00\x0b\xb0\x00\x02\x40\x00\x00\x00\x40\x00")}, PieceIndex: 0x00000bb0, BeginOffset: 0x00024000, PieceLength: 0x00004000}},
		{"Piece", "\x00\x00\x00\x0d\x07\x00\x00\x05\x2d\x00\x02\x80\x00\x11\x11\x11\x11",
			&PieceMessage{BasicMessage: BasicMessage{Type: MessageTypePiece, Length: 13, Payload: []byte("\x00\x00\x05\x2d\x00\x02\x80\x00\x11\x11\x11\x11")}, PieceIndex: 0x0000052d, BeginOffset: 0x00028000, Block: []byte("\x11\x11\x11\x11")}},
		{"Cancel", "\x00\x00\x00\x0d\x08\x00\x00\x05\x2d\x00\x02\x80\x00\x00\x00\x40\x00",
			&CancelMessage{BasicMessage: BasicMessage{Type: MessageTypeCancel, Length: 13, Payload: []byte("\x00\x00\x05\x2d\x00\x02\x80\x00\x00\x00\x40\x00")}, PieceIndex: 0x0000052d, BeginOffset: 0x00028000, PieceLength: 0x00004000}},
		{"Port", "\x00\x00\x00\x03\x09\xb9\xaa",
			&PortMessage{BasicMessage: BasicMessage{Type: MessageTypePort, Length: 3, Payload: []byte("\xb9\xaa")}, Port: 47530}},
	}

	Convey("Parsing messages from bytes", t, func() {
		for _, sm := range smTests {
			Convey(fmt.Sprintf("%s Message", sm.Desc), func() {
				br := bytes.NewReader([]byte(sm.String))
				m, err := ReadMessage(br)
				So(m, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(m, ShouldResemble, sm.Message)
			})
		}
	})

	Convey("Convert messages to bytes", t, func() {
		for _, sm := range smTests {
			Convey(fmt.Sprintf("%s Message", sm.Desc), func() {
				b := sm.Message.Bytes()
				So(b, ShouldResemble, []byte(sm.String))
			})
		}
	})
}
