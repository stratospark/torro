package structure

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTrackerRequest(t *testing.T) {
	Convey("Creating a Tracker Request struct", t, func() {
		filename := "../testfiles/TheInternetsOwnBoyTheStoryOfAaronSwartz_archive.torrent"
		metainfo := NewMetainfo(filename)
		request := NewTrackerRequest(metainfo)
		So(request, ShouldNotBeNil)
		So(request.InfoHash, ShouldEqual, "%29%eb%26%d6%ba%89d%9c%10%5d%c8%e2~%af%dc%0c.%f6%22%92")
		So(request.PeerID, ShouldEqual, "-TR2840-nj5ovtkoz2ed8")

	})
}
