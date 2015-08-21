package bencoding

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDecodeString(t *testing.T) {
	Convey("When given a valid input", t, func() {
		Convey("Decoding '4:blah' returns 'blah'.", func() {
			result, err := DecodeString("4:blah")
			So(result, ShouldEqual, "blah")
			So(err, ShouldBeNil)
		})

		Convey("Decoding '0:' returns empty string.", func() {
			result, err := DecodeString("0:")
			So(result, ShouldEqual, "")
			So(err, ShouldBeNil)
		})

		Convey("Decoding '10:asdfghjklq' works. 2 digit length.", func() {
			result, err := DecodeString("10:asdfghjklq")
			So(result, ShouldEqual, "asdfghjklq")
			So(err, ShouldBeNil)
		})

		Convey("Decoding '15:this:has:colons' works.", func() {
			result, err := DecodeString("15:this:has:colons")
			So(result, ShouldEqual, "this:has:colons")
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

type InputResultErr struct {
	input  interface{}
	result interface{}
	hasError bool
	err    error

}

func TestDecodeInteger(t *testing.T) {
	Convey("Given a valid input", t, func() {
		cases := map[string]InputResultErr{
			"'i3e' returns 3": InputResultErr{"i3e", 3, false, nil},
			"'i-1e' returns -1": InputResultErr{"i-1e", -1, false, nil},
			"'i100e' returns 100": InputResultErr{"i100e", 100, false, nil},
			"'i0e' returns 0": InputResultErr{"i0e", 0, false, nil},
		}
		for description, test := range cases {
			Convey(fmt.Sprintf("%s", description), func() {
				input := test.input.(string)
				result, err := DecodeInteger(input)
				So(result, ShouldEqual, test.result)
				So(err, ShouldEqual, test.err)
			})
		}
	})

	Convey("Given an invalid input", t, func() {
		cases := map[string]InputResultErr{
			"'i04e'": InputResultErr{"i04e", 0, true, ErrDecodeIntegerNoPadding},
			"'iae'": InputResultErr{"iae", 0, true, nil},
			"'e9a": InputResultErr{"e9a", 0, true, ErrDecodeIntegerBadFormat},
		}
		for description, test := range cases {
			Convey(fmt.Sprintf("%s", description), func() {
				input := test.input.(string)
				result, err := DecodeInteger(input)
				So(result, ShouldEqual, test.result)
				if test.hasError && test.err != nil {
				So(err, ShouldEqual, test.err)
				} else {
					So(err, ShouldNotBeNil)
				}
			})
		}
	})

}
