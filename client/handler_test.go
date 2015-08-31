package client

import (
	. "github.com/smartystreets/goconvey/convey"
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
}
