package packet

import (
	"fmt"
	"testing"
	"time"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
)

func TestPacketCoding(t *testing.T) {
	exampleIID := types.CodableList{}
	exampleIID.Add(1, types.NewCodableIID(1, 1, []int{}))
	exampleIID.Add(2, types.NewCodableIID(6, 4, []int{1}))
	exampleIID.Add(3, types.NewCodableIID(6, 2, []int{1, 2}))
	p := NewGetRequestPacket(exampleIID)
	p.timestamp = types.NewCodableTimestamp(time.Date(2024, 7, 8, 23, 0, 15, 152000000, time.UTC))
	p.messageId = "NEE6QSYZ28R520a3"
	encoded := p.Encode()
	// if encoded != "kdk847ufh84jg87g\x00GT\x007\x008\x007\x002024\x0023\x000\x0015\x00152\x00NEE6QSYZ28R520a3\x003\x00D\x002\x001\x001\x00D\x003\x006\x004\x001\x00D\x004\x006\x002\x001\x002\x000\x000\x00" {
	// 	t.Errorf("Error in Encoding Packet")
	// }
	p1 := &LSNMPvS_Packet{}
	_, err := p1.Decode(encoded)
	if err != 0 {
		t.Errorf(err.Error())
	}
	if !p1.Equal(p) {
		t.Errorf("Error in Decoding Packet")
	}
	fmt.Println(p)
	fmt.Println(p1)
}


func TestEncriprion(t *testing.T) {
	text := "Hello, World!"
	encrypted := Encrypt(text)
	decrypted := Decrypt(encrypted)
	if text != decrypted {
		t.Errorf("Error in Encryption")
	}
}