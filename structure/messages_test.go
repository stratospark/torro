package structure

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHandshake(t *testing.T) {
	Convey("Parse a handshake message from bytes", t, func() {
		msg := "\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		br := bytes.NewReader([]byte(msg))
		hs, err := NewHandshake(br)
		So(hs, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})

	// "Reject a malformed handshake message"
	Convey("Parse a handshake message from bytes", t, func() {
		msg := "\x13\x41\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		br := bytes.NewReader([]byte(msg))
		hs, err := NewHandshake(br)
		So(hs, ShouldBeNil)
		So(err, ShouldEqual, ErrNotBitTorrentProtocol)
	})

	Convey("Convert a handshake message to bytes", t, func() {
		msg := "\x13\x42\x69\x74\x54\x6f\x72\x72\x65\x6e\x74\x20\x70\x72\x6f\x74\x6f\x63\x6f\x6c\x00\x00\x00\x00\x00\x10\x00\x05\x6f\xda\xb6\xc1\x9f\x72\x14\x76\xfa\xca\xab\x36\x60\x8a\x87\x7a\x2a\xac\xbf\xc9\x2d\x55\x54\x33\x34\x34\x30\x2d\xcf\x9f\x51\x2b\xce\x01\x31\xf9\x38\x6f\xb6\x98"
		br := bytes.NewReader([]byte(msg))
		hs, _ := NewHandshake(br)
		b := hs.Bytes()
		So(b, ShouldNotBeNil)
		So(b, ShouldResemble, []byte(msg))
	})
}

type StringMessageTest struct {
	String  string
	Message Message
}

func TestMessages(t *testing.T) {

	smTests := []StringMessageTest{
		{"\x00\x00\x00\x00",
			&BasicMessage{Type: MessageTypeKeepAlive, Length: 0}},
		{"\x00\x00\x00\x01\x00",
			&BasicMessage{Type: MessageTypeChoke, Length: 1}},
		{"\x00\x00\x00\x01\x01",
			&BasicMessage{Type: MessageTypeUnchoke, Length: 1}},
		{"\x00\x00\x00\x01\x02",
			&BasicMessage{Type: MessageTypeInterested, Length: 1}},
		{"\x00\x00\x00\x01\x03",
			&BasicMessage{Type: MessageTypeNotInterested, Length: 1}},
		{"\x00\x00\x00\x05\x04\x00\x00\x18\xa4",
			&BasicMessage{Type: MessageTypeHave, Length: 5, Payload: []byte("\x00\x00\x18\xa4")}},
	}

	for _, sm := range smTests {
		Convey("Parse an 'Interested' message from bytes", t, func() {
			br := bytes.NewReader([]byte(sm.String))
			m, err := ReadMessage(br)
			So(m, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(m, ShouldResemble, sm.Message)
		})

	}

	Convey("Convert an 'Interested' message to bytes", t, func() {
		m := &BasicMessage{Length: 1, Type: MessageTypeInterested}
		b := m.Bytes()
		msg := "\x00\x00\x00\x01\x02"
		So(b, ShouldResemble, []byte(msg))
	})
}
