package structure

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestMetainfo(t *testing.T) {

	Convey("Should properly serialize torrent file into Metainfo structure", t, func() {
		filename := "../ubuntu.torrent"
		metainfo := NewMetainfo(filename)
		So(metainfo, ShouldNotBeNil)
		So(metainfo.Announce, ShouldEqual, "http://torrent.ubuntu.com:6969/announce")

		loc, _ := time.LoadLocation("US/Pacific")
		So(metainfo.CreationDate, ShouldHappenWithin, time.Duration(0), time.Date(2014, 7, 24, 16, 52, 15, 0, loc))

		So(metainfo.Comment, ShouldEqual, "Ubuntu CD releases.ubuntu.com")
	})
}
