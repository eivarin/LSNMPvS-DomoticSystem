package CodableValues

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	NullCharStr = string('\x00')
	intPtrn	 = `^(\d+)\x00(.*)$`
	
)

func EncodeInt(value int) string {
	return strconv.Itoa(value) + NullCharStr
}

func DecodeInt(data string) (int, string, error) {
	splitted := strings.SplitN(data, NullCharStr, 2)
	if len(splitted) == 2 {
		res, err := strconv.Atoi(splitted[0])
		if err != nil {
			return 0, "", err
		}
		return res, splitted[1], nil
	}
	return 0, "", nil
}

func EncodeInt64(value int64) string {
	return strconv.FormatInt(value, 10) + NullCharStr
}

func DecodeInt64(data string) (int64, string, error) {
	splitted := strings.SplitN(data, NullCharStr, 2)
	if len(splitted) == 2 {
		res, err := strconv.ParseInt(splitted[0], 10, 64)
		if err != nil {
			return 0, "", err
		}
		return res, splitted[1], nil
	}
	return 0, "", nil
}

func EncodeByte(value byte) string {
	return string(value) + NullCharStr
}

func DecodeByte(data string) (byte, string, error) {
	res := byte(0)
	rest := ""
	if len(data) > 0  && data[1] == NullCharStr[0] {
		res = data[0]
		rest = data[1:]
		return res, rest, nil
	}
	return res, rest, fmt.Errorf("invalid data")
}

func EncodeString(value string) string {
	return value + NullCharStr
}

func DecodeString(data string) (string, string, error) {
	res := ""
	rest := ""
	splitted := strings.SplitN(data, NullCharStr, 2)
	if len(splitted) == 2 {
		res = splitted[0]
		rest = splitted[1]
	}
	return res, rest, nil
}
