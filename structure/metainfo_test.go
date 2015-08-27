package structure

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func xTestMetainfo(t *testing.T) {

	Convey("Should properly serialize torrent file into Metainfo structure", t, func() {

		Convey("Given a Single File Mode torrent", func() {
			filename := "../testfiles/ubuntu.torrent"
			metainfo := NewMetainfo(filename)
			So(metainfo, ShouldNotBeNil)
			So(metainfo.Announce, ShouldEqual, "http://torrent.ubuntu.com:6969/announce")

			loc, _ := time.LoadLocation("US/Pacific")
			So(metainfo.CreationDate, ShouldHappenWithin, time.Duration(0), time.Date(2014, 7, 24, 16, 52, 15, 0, loc))

			So(metainfo.Comment, ShouldEqual, "Ubuntu CD releases.ubuntu.com")

			So(metainfo.Info.PieceLength, ShouldEqual, 524288)
			So(len(metainfo.Info.Pieces), ShouldEqual, 39240)
			So(metainfo.Info.Mode, ShouldEqual, InfoModeSingle)

			So(metainfo.Info.Name, ShouldEqual, "ubuntu-14.04.1-desktop-amd64.iso")
			So(metainfo.Info.Length, ShouldEqual, 1028653056)
		})
	})
}
