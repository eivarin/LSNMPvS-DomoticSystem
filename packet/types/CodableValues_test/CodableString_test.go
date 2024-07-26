package CodableValues_test

import (
	"testing"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

func TestCodableStringEncoding(t *testing.T) {
	cvs := &CodableValues.CodableString{Value: "test"}
	if cvs.Encode() != "test\x00" {
		t.Errorf("Error in CodableString.Encode()")
	}
}

func TestCodableStringDecoding(t *testing.T) {
	cvs := &CodableValues.CodableString{}
	l := 5
	cvs.Decode("test\x00", &l)
	if cvs.Value != "test" {
		t.Errorf("Error in CodableString.Decode()")
	}
}
