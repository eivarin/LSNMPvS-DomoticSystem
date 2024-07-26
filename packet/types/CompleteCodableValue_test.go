package types

import (
	"testing"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

func TestCompleteIIDSimple(t *testing.T) {
	ciid := NewCodableIID(6, 0, []int{})
	encoded := ciid.Encode()
	if encoded != "D\x002\x006\x000\x00" {
		t.Errorf("Error in Encoding simple IID")
	}
	ciid2 := &CompleteCodableValue{}
	ciid2.Decode(encoded)
	iid := ciid2.Value.(*CodableValues.IID)
	if ciid2.DataType != 'D' || ciid2.Length != 2 || iid.Structure != 6 || iid.Object != 0 || iid.FirstIndex != nil || iid.SecondIndex != nil {
		t.Errorf("Error in Decoding simple IID")
	}
}