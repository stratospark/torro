package structure

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTrackerRequest(t *testing.T) {
	Convey("Creating a Tracker Request struct", t, func() {
		request := NewTrackerRequest()
		So(request, ShouldNotBeNil)
	})
}
