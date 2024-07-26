package CodableValues_test

import (
	"testing"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

func TestIIDSimpleEncoding(t *testing.T) {
    iid := CodableValues.NewIID(6, 0)
	if iid.Encode() != "6\x000\x00" {
		t.Errorf("Error in Encoding simple IID")
	}
}

func TestIIDSingleIndexEncoding(t *testing.T) {
	iid := CodableValues.NewIIDSingleIndex(6, 1, 2)
	if iid.Encode() != "6\x001\x002\x00" {
		t.Errorf("Error in Encoding single index IID")
	}
}

func TestIIDDoubleIndexEncoding(t *testing.T) {
	iid := CodableValues.NewIIDDoubleIndex(6, 1, 2, 5)
	if iid.Encode() != "6\x001\x002\x005\x00" {
		t.Errorf("Error in Encoding double index IID")
	}
}

func TestIIDSimpleDecoding(t *testing.T) {
	iid := new(CodableValues.IID)
	l := 2
	iid.Decode("6\x000\x00", &l)
	if iid.Structure != 6 || iid.Object != 0 || iid.Length != 2 {
		t.Errorf("Error in Decoding simple IID")
	}
}

func TestIIDSingleIndexDecoding(t *testing.T) {
	iid := new(CodableValues.IID)
	l := 3
	iid.Decode("6\x001\x002\x00", &l)
	if iid.Structure != 6 || iid.Object != 1 || *iid.FirstIndex != 2 || iid.Length != 3 {
		t.Errorf("Error in Decoding single index IID")
	}
}

func TestIIDDoubleIndexDecoding(t *testing.T) {
	iid := new(CodableValues.IID)
	l := 4
	iid.Decode("6\x001\x002\x005\x00", &l)
	if iid.Structure != 6 || iid.Object != 1 || *iid.FirstIndex != 2 || *iid.SecondIndex != 5 || iid.Length != 4 {
		t.Errorf("Error in Decoding double index IID")
	}
}