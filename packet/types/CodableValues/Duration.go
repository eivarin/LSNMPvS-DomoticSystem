package CodableValues

import (
	"time"
)

type Duration struct {
	Value time.Duration
}

func NewDuration(value time.Duration) *Duration {
	return &Duration{Value: value}
}

func (cvd *Duration) Encode() string {
	trunc := []time.Duration{
		24 * time.Hour,
		time.Hour,
		time.Minute,
		time.Second,
		time.Millisecond,
	}
	total := cvd.Value
	enc := ""
	for _, t := range trunc {
		truncated := total.Truncate(t)
		total -= truncated
		enc += EncodeInt(int(truncated / t))
	}
	return enc
}

func (cvd *Duration) Decode(data string, length *int) (string, error) {
	days, rest, err := DecodeInt64(data)
	if err != nil {
		return "", err
	}
	hours, rest, err := DecodeInt64(rest)
	if err != nil {
		return "", err
	}
	minutes, rest, err := DecodeInt64(rest)
	if err != nil {
		return "", err
	}
	seconds, rest, err := DecodeInt64(rest)
	if err != nil {
		return "", err
	}
	miliseconds, rest, err := DecodeInt64(rest)
	if err != nil {
		return "", err
	}
	total := time.Duration(0)
	total += time.Duration(days) * 24 * time.Hour
	total += time.Duration(hours) * time.Hour
	total += time.Duration(minutes) * time.Minute
	total += time.Duration(seconds) * time.Second
	total += time.Duration(miliseconds) * time.Millisecond
	cvd.Value = total
	return rest, nil
}

func (cvd *Duration) Equals(other interface{}) bool {
	ActualValue := other.(*Duration)
	return cvd.Value == ActualValue.Value
}

func (cvd *Duration) String() string {
	return "Dur{" + cvd.Value.String() + "}"
}

func (cvd *Duration) Copy() *Duration {
	return NewDuration(cvd.Value)
}
