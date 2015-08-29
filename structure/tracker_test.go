package structure

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"strconv"
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

	Convey("Getting a GET /announce URL", t, func() {
		filename := "../testfiles/kali-linux-2.0-i386.iso.torrent"
		metainfo := NewMetainfo(filename)
		request := NewTrackerRequest(metainfo)
		So(request, ShouldNotBeNil)

		url := metainfo.Announce +
		"?info_hash=" + metainfo.Info.Hash +
		"&peer_id=" + request.PeerID +
		"&uploaded=" + strconv.Itoa(request.Uploaded) +
		"&downloaded=" + strconv.Itoa(request.Downloaded) +
		"&left=" + strconv.Itoa(metainfo.Info.TotalBytes)

		So(request.GetURL(), ShouldEqual, url)
	})

}


/*
http://linuxtracker.org:2710/00000000000000000000000000000000/announce?info_hash=o%da%b6%c1%9fr%14v%fa%ca%ab6%60%8a%87z*%ac%bf%c9&peer_id=-qB3230-J1B5U1f7viyy&port=8999&uploaded=0&downloaded=0&left=3403579459&corrupt=0&key=6DC4E096&event=started&numwant=200&compact=1&no_peer_id=1&supportcrypto=1&redundant=0
 */
