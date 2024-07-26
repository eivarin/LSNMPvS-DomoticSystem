package CodableValues_test

import (
	"testing"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

func TestCodableIntEncoding(t *testing.T) {
	cvi := &CodableValues.CodableInt{Value: 56}
	if cvi.Encode() != "56\x00" {
		t.Errorf("Error in CodableInt.Encode()")
	}
}

func TestCodableIntDecoding(t *testing.T) {
	cvi := &CodableValues.CodableInt{}
	l := 2
	cvi.Decode("56\x00", &l)
	if cvi.Value != 56 {
		t.Errorf("Error in CodableInt.Decode()")
	}
}
