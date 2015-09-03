package structure

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBitfield(t *testing.T) {
	Convey("Create bitfield of arbitrary length", t, func() {
		h := "\xff\x01"

		bf := BitFieldFromHexString(h)
		So(bf.Bytes(), ShouldResemble, []byte("\xff\x01"))
		So(bf.String(), ShouldEqual, "1111111100000001")
		So(bf.Get(10), ShouldEqual, 0)

		bf.Set(10, 1)
		So(bf.Get(10), ShouldEqual, 1)
		So(bf.Bytes(), ShouldResemble, []byte("\xff\x21"))
		So(bf.String(), ShouldEqual, "1111111100100001")
	})
}
