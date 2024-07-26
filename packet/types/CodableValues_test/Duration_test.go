package CodableValues_test

import (
	"testing"
	"time"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

func TestDurationCoding(t *testing.T) {
	t1 := time.Date(2024, 7, 8, 20, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 7, 8, 23, 0, 15, 0, time.UTC)
	delta := t2.Sub(t1)
	d := CodableValues.NewDuration(delta)
	if d.Encode() != "0\x003\x000\x0015\x000\x00" {
		t.Errorf("Error in Encoding Duration")
	}
	d1 := new(CodableValues.Duration)
	d1.Decode("0\x003\x000\x0015\x000\x00", nil)
	if d1.Value != delta {
		t.Errorf("Error in Decoding Duration")
	}
}