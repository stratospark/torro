package encoding

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDecodeString(t *testing.T) {
	Convey("When given a valid input", t, func() {
		Convey("Decoding '4:blah' should return 'blah'.", func() {
			result, err := DecodeString("4:blah")
			So(result, ShouldEqual, "blah")
			So(err, ShouldBeNil)
		})

		Convey("Decoding '0:' should return ''", func() {
			result, err := DecodeString("0:")
			So(result, ShouldEqual, "")
			So(err, ShouldBeNil)
		})
	})

	Convey("Given an invalid input", t, func() {
		Convey("Decoding '4:wrong' should return err BadLength", func() {
			result, err := DecodeString("4:wrong")
			So(result, ShouldEqual, "")
			So(err, ShouldEqual, ErrDecodeStringBadLength)
		})

		Convey("Decoding 'a:wrong' should return err", func() {
			result, err := DecodeString("a:wrong")
			So(result, ShouldEqual, "")
			So(err, ShouldNotBeNil)
		})

	})
}
