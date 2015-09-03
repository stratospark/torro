package structure

import (
	"bytes"
	"fmt"
	"math"
)

var _ = math.Ceil
var _ = fmt.Printf

/**
Bitmap algorithm. (BitSet With mulit-choiced bit width.)

Most of the code taken from: https://github.com/smalllixin/bitarray/blob/master/bitarr.go
*/

type BitArray struct {
	B             []byte //we use bit
	valueBitWidth byte   // how many bit present one value. only value 1,2,4,8 is supported
	countPerByte  byte
	bitmapLen     uint32
}

/*
bitmapLen: how many bit we should save
*/
func NewBitArray(bitmapLen uint32, valueBitWidth byte) *BitArray {
	return new(BitArray).Init(bitmapLen, valueBitWidth)
}

func (s *BitArray) Init(bitmapLen uint32, valueBitWidth byte) *BitArray {
	validBitLen := false
	for i := uint(0); i < 4; i++ {
		if valueBitWidth == 0x08>>i {
			validBitLen = true
			s.countPerByte = 0x08 / valueBitWidth
			break
		}
	}
	if !validBitLen {
		panic("BitArray validBitLen only 1,2,4,8 is supported")
	}
	s.B = make([]byte, bitmapLen/uint32(s.countPerByte)+1)
	s.bitmapLen = bitmapLen
	s.valueBitWidth = valueBitWidth
	return s
}

func (s *BitArray) GetAllocLen() int {
	return len(s.B)
}

func (s *BitArray) SetB(pos uint32, val byte) {
	whichByte := pos / uint32(s.countPerByte)
	whichPos := pos % uint32(s.countPerByte)
	n := byte(whichPos)
	w := s.valueBitWidth
	oo := (byte(0xFF<<(8-w)) >> (n * w)) ^ 0xFF
	zr := s.B[whichByte] & oo         //something like [rr00 rrrr]
	sr := byte(val<<(8-w)) >> (n * w) // [00ss 0000]
	s.B[whichByte] = zr | sr
}

func (s *BitArray) GetBytes() []byte {
	return s.B
}

func (s *BitArray) GetB(pos uint32) byte {
	whichByte := pos / uint32(s.countPerByte)
	whichPos := pos % uint32(s.countPerByte)
	n := byte(whichPos)
	w := s.valueBitWidth

	oo := (byte(0xFF<<(8-w)) >> (n * w)) // 0011 0000
	oorr := s.B[whichByte] & oo          //00rr 0000
	return oorr >> (8 - (n+1)*w)
}

type BitField struct {
	ba *BitArray
}

func BitFieldFromHexString(hs string) *BitField {
	//t.Log(len(h))
	bm := NewBitArray(uint32(len(hs)*8), 1)
	bm.B = []byte(hs)
	return &BitField{ba: bm}
}

func (bf *BitField) Set(pos uint32, val byte) {
	bf.ba.SetB(pos, val)
}

func (bf *BitField) Get(pos uint32) byte {
	return bf.ba.GetB(pos)
}

func (bf *BitField) Bytes() []byte {
	return bf.ba.GetBytes()
}

func (bf *BitField) String() string {
	var buf bytes.Buffer
	for _, b := range bf.Bytes() {
		fmt.Fprintf(&buf, "%08b", b)
	}
	return fmt.Sprintf("%s", buf.Bytes())
}
