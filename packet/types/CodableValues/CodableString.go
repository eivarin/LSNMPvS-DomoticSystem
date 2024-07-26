package CodableValues

type CodableString struct {
	Value string
}

func (cvs *CodableString) Encode() string {
	return EncodeString(cvs.Value)
}

func (cvs *CodableString) Decode(data string, length *int) (string, error) {
	value, rest, err := DecodeString(data)
	if err != nil {
		return "", err
	}
	cvs.Value = value
	return rest, nil
}

func (cvs *CodableString) Equals(other interface{}) bool {
	ActualValue := other.(*CodableString)
	return cvs.Value == ActualValue.Value
}

func (cvs *CodableString) String() string {
	return cvs.Value
}

func (cvs *CodableString) Copy() *CodableString {
	return &CodableString{Value: cvs.Value}
}