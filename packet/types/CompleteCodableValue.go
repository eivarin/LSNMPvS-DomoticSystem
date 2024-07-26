package types

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type CompleteCodableValue struct {
	DataType byte
	Length   int
	Value    CodableValueI
}

func NewCodableInt(value int) *CompleteCodableValue {
	return &CompleteCodableValue{
		DataType: 'I',
		Length:   1,
		Value:    &CodableValues.CodableInt{Value: value},
	}
}

func NewCodableString(value string) *CompleteCodableValue {
	return &CompleteCodableValue{
		DataType: 'S',
		Length:   1,
		Value:    &CodableValues.CodableString{Value: value},
	}
}

func NewCodableTimestamp(ts time.Time) *CompleteCodableValue {
	return &CompleteCodableValue{
		DataType: 'T',
		Length:   7,
		Value:    CodableValues.NewTimestamp(ts),
	}
}

func NewCodableTimestampNow() *CompleteCodableValue {
	return &CompleteCodableValue{
		DataType: 'T',
		Length:   7,
		Value:    CodableValues.NewTimestampNow(),
	}
}

func NewCodableDuration(d time.Duration) *CompleteCodableValue {
	return &CompleteCodableValue{
		DataType: 'T',
		Length:   5,
		Value:    CodableValues.NewDuration(d),
	}
}

func NewCodableIID(Structure, Object int, Indexes []int) *CompleteCodableValue {
	l := 2 + len(Indexes)
	var iid *CodableValues.IID
	switch l {
	case 2:
		iid = CodableValues.NewIID(Structure, Object)
	case 3:
		iid = CodableValues.NewIIDSingleIndex(Structure, Object, Indexes[0])
	case 4:
		iid = CodableValues.NewIIDDoubleIndex(Structure, Object, Indexes[0], Indexes[1])
	default:
		return nil
	}
	return &CompleteCodableValue{
		DataType: 'D',
		Length:   l,
		Value:    iid,
	}
}

func (cvd *CompleteCodableValue) Encode() string {
	return CodableValues.EncodeByte(cvd.DataType) + CodableValues.EncodeInt(cvd.Length) + cvd.Value.Encode()
}

func (cvd *CompleteCodableValue) Decode(data string) (string, error) {
	var unparsedValue string
	ptrn, _ := regexp.Compile(`^([ISTD])\x00(\d+)\x00(.+)$`)
	matches := ptrn.FindStringSubmatch(data)
	cvd.DataType = matches[1][0]
	cvd.Length, _ = strconv.Atoi(matches[2])
	unparsedValue = matches[3]
	switch cvd.DataType {
	case 'I':
		cvd.Value = &CodableValues.CodableInt{}
	case 'S':
		cvd.Value = &CodableValues.CodableString{}
	case 'T':
		if cvd.Length == 7 {
			cvd.Value = &CodableValues.Timestamp{}
		} else if cvd.Length == 5 {
			cvd.Value = &CodableValues.Duration{}
		} else {
			return "", fmt.Errorf("invalid length for Timestamp or Duration")
		}
	case 'D':
		cvd.Value = &CodableValues.IID{}
	default:
		return "", fmt.Errorf("invalid data type")
	}
	rest, error := cvd.Value.Decode(unparsedValue, &cvd.Length)
	if error != nil {
		return "", error
	}
	return rest, nil
}

func (cvd *CompleteCodableValue) Equals(other interface{}) bool {
	ActualValue := other.(*CompleteCodableValue)
	return cvd.DataType == ActualValue.DataType && cvd.Length == ActualValue.Length && cvd.Value.Equals(ActualValue.Value)
}

func (cvd *CompleteCodableValue) String() string {
	return cvd.Value.String()
}

func (cvd *CompleteCodableValue) Copy() *CompleteCodableValue {
	valueCopy := CopyCodableValueI(cvd.Value)
	r := &CompleteCodableValue{
		DataType: cvd.DataType,
		Length:   cvd.Length,
		Value:    valueCopy,
	}
	return r
}
