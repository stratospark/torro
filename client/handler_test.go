package client

import (
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	Convey("Listens to incoming connections on a given port", t, func() {
		port := 55555
		s := NewBTService(port)
		s.StartListening()

		time.Sleep(time.Millisecond)
		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		_ = s.StopListening()
		time.Sleep(time.Millisecond)
		So(s.Listening, ShouldBeFalse)
	})

	Convey("Accepts a handshake and adds to the connection list", t, func() {
		port := 55555
		s := NewBTService(port)
		s.StartListening()

		time.Sleep(time.Millisecond)
		So(s.Listener, ShouldNotBeNil)
		So(s.Listening, ShouldBeTrue)

		addr, _ := net.ResolveTCPAddr("tcp", "localhost:55555")
		conn, err := net.DialTCP("tcp", nil, addr)
		So(err, ShouldBeNil)

		conn.Write([]byte("ping"))
		buf := make([]byte, 4)
		_, err = io.ReadFull(conn, buf)
		t.Log(buf)
		So(err, ShouldBeNil)
		So(buf, ShouldResemble, []byte("pong"))

		_ = s.StopListening()
		time.Sleep(time.Millisecond)
		So(s.Listening, ShouldBeFalse)
	})
}
