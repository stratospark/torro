package structure

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestMetainfo(t *testing.T) {

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

			So(metainfo.Info.Files, ShouldBeNil)
		})

		Convey("Given a Multiple File Mode torrent", func() {
			filename := "../testfiles/TheInternetsOwnBoyTheStoryOfAaronSwartz_archive.torrent"
			metainfo := NewMetainfo(filename)
			So(metainfo, ShouldNotBeNil)

			So(metainfo.Info.Files, ShouldNotBeNil)

			file0 := metainfo.Info.Files[0]
			So(file0.Length, ShouldEqual, 1466)
			So(file0.MD5sum, ShouldEqual, "8969eabd433acad882bc994b21ecc9b4")
			So(file0.Path, ShouldEqual, "TheInternetsOwnBoyTheStoryOfAaronSwartz_meta.xml")

			file1 := metainfo.Info.Files[1]
			So(file1.Length, ShouldEqual, 4192838)
			So(file1.Path, ShouldEqual, ".____padding_file/0")

			So(metainfo.Info.Hash, ShouldEqual, "%29%eb%26%d6%ba%89d%9c%10%5d%c8%e2~%af%dc%0c.%f6%22%92")
		})
	})
}
