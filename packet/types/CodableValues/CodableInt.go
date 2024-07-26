package CodableValues

import "strconv"


type CodableInt struct {
	Value int
}

func (cvi *CodableInt) Encode() string {
	return EncodeInt(cvi.Value)
}

func (cvi *CodableInt) Decode(data string, length *int) (string, error) {
	value, rest, err := DecodeInt(data)
	if err != nil {
		return "", err
	}
	cvi.Value = value
	return rest, nil
}

func (cvi *CodableInt) Equals(other interface{}) bool {
	ActualValue	:= other.(*CodableInt)
	return cvi.Value == ActualValue.Value
}

func (cvi *CodableInt) String() string {
	return strconv.Itoa(cvi.Value)
}

func (cvi *CodableInt) Copy() *CodableInt {
	return &CodableInt{Value: cvi.Value}
}
