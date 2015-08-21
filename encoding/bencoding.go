package encoding
import (
	"strings"
	"strconv"
	"errors"
)

var ErrDecodeStringBadLength = errors.New("Given length does not match length of string")

func DecodeString(s string) (result string, err error) {
	parts := strings.SplitN(s, ":", -1)
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