package netfuncs

import (
	"net"
	"strconv"
)

func Send(udpAddr *net.UDPAddr, message []byte) error {
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(message)
	if err != nil {
		return err
	}
	return nil
}

func SendStrAddr(address string, message []byte) error {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	return Send(udpAddr, message)
}

func SendBroadcast(port int, message []byte) error {
	udpAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	return Send(udpAddr, message)
}