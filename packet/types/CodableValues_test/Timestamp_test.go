package CodableValues_test

import (
	"testing"
	"time"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

func TestTimeStampCoding(t *testing.T) {
	toCheck := time.Date(2024, 7, 8, 23, 0, 15, 152000000, time.UTC)
	ts := CodableValues.NewTimestamp(toCheck)
	encoded := ts.Encode()
	if encoded != "8\x007\x002024\x0023\x000\x0015\x00152\x00" {
		t.Errorf("Error in Encoding Timestamp")
	}
	ts1 := new(CodableValues.Timestamp)
	ts1.Decode("8\x007\x002024\x0023\x000\x0015\x00152\x00", nil)
	if !ts1.Ts.Equal(toCheck) {
		t.Errorf("Error in Decoding Timestamp")
	}
}