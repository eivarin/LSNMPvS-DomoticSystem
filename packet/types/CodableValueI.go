package types

import "github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"

type CodableValueI interface {
	Encode() string
	Decode(string, *int) (string, error)
	Equals(interface{}) bool
	String() string
}

func CopyCodableValueI(ci CodableValueI) CodableValueI {
	var valueCopy CodableValueI
	switch ciTrue := ci.(type) {
	case *CodableValues.IID:
		valueCopy = ciTrue.Copy()
	case *CodableValues.CodableInt:
		valueCopy = ciTrue.Copy()
	case *CodableValues.CodableString:
		valueCopy = ciTrue.Copy()
	case *CodableValues.Timestamp:
		valueCopy = ciTrue.Copy()
	case *CodableValues.Duration:
		valueCopy = ciTrue.Copy()
	}
	return valueCopy
}