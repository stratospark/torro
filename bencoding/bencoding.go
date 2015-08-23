package bencoding

import (
	"errors"
	"strconv"
	"strings"
)

var ErrDecodeStringBadLength = errors.New("Given length does not match length of string.")
var ErrDecodeIntegerBadFormat = errors.New("Integer must begin with 'i' and end with 'e'.")
var ErrDecodeIntegerNoPadding = errors.New("Integer must not be padded with zeroes (0).")

func DecodeString(s string) (result string, rest string, err error) {
	parts := strings.SplitN(s, ":", 2)
	length, err := strconv.ParseInt(parts[0], 10, 32)
	rest = ""

	if err != nil {
		return "", rest, err
	}
	if length == 0 {
		result = ""
	} else {
		if len(parts[1]) < int(length) {
			result = ""
			err = ErrDecodeStringBadLength
		} else {
			result = parts[1][0:length]
			rest = parts[1][length:len(parts[1])]
		}
	}
	return result, rest, err
}

func DecodeInteger(s string) (result int64, err error) {
	if strings.HasPrefix(s, "i") && strings.HasSuffix(s, "e") {
		integerString := s[1 : len(s)-1]
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

//func DecodeList(s string) (result interface{}, err error) {
//	if strings.HasPrefix(s, "l") && strings.HasSuffix(s, "e") {
//		listString := s[1:len(s)-1]
//		switch listString[0] {
//		case 'i':
//			endIndex = strings.
//		}
//	}
//	return result, nil
//}
