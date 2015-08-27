package structure
import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTrackerRequest(t *testing.T) {
	Convey("Creating a Tracker Request struct", t, func() {
		request := NewTrackerRequest()
		So(request, ShouldNotBeNil)
	})
}
