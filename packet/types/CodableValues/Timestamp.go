package CodableValues

import (
	"fmt"
	"time"
)

const MiliToNano = 1000000

type Timestamp struct {
	Ts time.Time
}

func NewTimestamp(value time.Time) *Timestamp {
	return &Timestamp{Ts: value}
}

func NewTimestampNow() *Timestamp {
	return &Timestamp{Ts: time.Now()}
}

func (cvts *Timestamp) Encode() string {
	res := ""
	res += fmt.Sprintf("%d\x00%d\x00%d\x00%d\x00%d\x00%d\x00%d\x00", cvts.Ts.Day(), cvts.Ts.Month(), cvts.Ts.Year(), cvts.Ts.Hour(), cvts.Ts.Minute(), cvts.Ts.Second(), cvts.Ts.Nanosecond()/MiliToNano)
	return res
}

func (cvts *Timestamp) Decode(data string, length *int) (string, error) {
	rest := ""
	day, rest, err := DecodeInt(data)
	if err != nil {
		return "", err
	}
	month, rest, err := DecodeInt(rest)
	if err != nil {
		return "", err
	}
	year, rest, err := DecodeInt(rest)
	if err != nil {
		return "", err
	}
	hour, rest, err := DecodeInt(rest)
	if err != nil {
		return "", err
	}
	minute, rest, err := DecodeInt(rest)
	if err != nil {
		return "", err
	}
	second, rest, err := DecodeInt(rest)
	if err != nil {
		return "", err
	}
	milisecond, rest, err := DecodeInt(rest)
	if err != nil {
		return "", err
	}
	value := time.Date(year, time.Month(month), day, hour, minute, second, milisecond*MiliToNano, time.UTC)
	cvts.Ts = value
	return rest, nil
}

func (cvts *Timestamp) String() string {
	return "Ts{" + cvts.Ts.Format("02/01/2006 15:04:05.9999999.") + "}"
}


func (cvts *Timestamp) Equals(other interface{}) bool {
	ActualValue := other.(*Timestamp)
	return cvts.Ts.Equal(ActualValue.Ts)
}

func (cvts *Timestamp) Copy() *Timestamp {
	return NewTimestamp(cvts.Ts)
}