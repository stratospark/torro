package bencoding

import (
	"errors"
	"strconv"
	"strings"
)

var ErrDecodeStringBadLength = errors.New("Given length does not match length of string.")
var ErrDecodeIntegerBadFormat = errors.New("Integer must begin with 'i' and end with 'e'.")
var ErrDecodeIntegerNoPadding = errors.New("Integer must not be padded with zeroes (0).")

func DecodeString(s string) (result string, err error) {
	parts := strings.SplitN(s, ":", 2)
	length, err := strconv.ParseInt(parts[0], 10, 32)

	if err != nil {
		return "", err
	}
	if length == 0 {
		result = ""
	} else {
		result = parts[1]
		if len(result) != int(length) {
			result = ""
			err = ErrDecodeStringBadLength
		}
	}
	return result, err
}

func DecodeInteger(s string) (result int64, err error) {
	if strings.HasPrefix(s, "i") && strings.HasSuffix(s, "e") {
		integerString := s[1:len(s)-1]
		if strings.HasPrefix(integerString, "0") && len(integerString) > 1 {
			return 0, ErrDecodeIntegerNoPadding
		}
		result, err = strconv.ParseInt(integerString, 10, 32)
		if err != nil {
			return 0, err
		}
	} else {
		result = 0
		err = ErrDecodeIntegerBadFormat
	}
	return result, err
}

