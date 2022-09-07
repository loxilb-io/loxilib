// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"encoding/binary"
	"net"
	"unsafe"
	"syscall"
	"errors"
	"bytes"
)

// Ntohl - Network to host byte-order long
func Ntohl(i uint32) uint32 {
	return binary.BigEndian.Uint32((*(*[4]byte)(unsafe.Pointer(&i)))[:])
}

// Htonl - Host to network byte-order long
func Htonl(i uint32) uint32 {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return *(*uint32)(unsafe.Pointer(&b[0]))
}

// Htons - Host to network byte-order short
func Htons(i uint16) uint16 {
	var j = make([]byte, 2)
	binary.BigEndian.PutUint16(j[0:2], i)
	return *(*uint16)(unsafe.Pointer(&j[0]))
}

// Ntohs - Network to host byte-order short
func Ntohs(i uint16) uint16 {
	return binary.BigEndian.Uint16((*(*[2]byte)(unsafe.Pointer(&i)))[:])
}

// IPtonl - Convert net.IP to network byte-order long  
func IPtonl(ip net.IP) uint32 {
	var val uint32

	if len(ip) == 16 {
		val = uint32(ip[12])
		val |= uint32(ip[13]) << 8
		val |= uint32(ip[14]) << 16
		val |= uint32(ip[15]) << 24
	} else {
		val = uint32(ip[0])
		val |= uint32(ip[1]) << 8
		val |= uint32(ip[2]) << 16
		val |= uint32(ip[3]) << 24
	}

	return val
}

// NltoIP - Convert network byte-order long to net.IP
func NltoIP(addr uint32) net.IP {
	var dip net.IP

	dip = append(dip, uint8(addr&0xff))
	dip = append(dip, uint8(addr>>8&0xff))
	dip = append(dip, uint8(addr>>16&0xff))
	dip = append(dip, uint8(addr>>24&0xff))

	return dip
}

// ArpPing - sends a arp request given the DIP, SIP and interface name
func ArpPing(DIP net.IP, SIP net.IP, ifName string) (int, error) {
	bZeroAddr := []byte{ 0x0, 0x0, 0x0, 0x0, 0x0, 0x0 }
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, int(Htons(syscall.ETH_P_ARP)))
	if err != nil {
		return -1, errors.New("af-packet-err")
	}
	defer syscall.Close(fd)

	if err := syscall.BindToDevice(fd, ifName); err != nil {
		return -1, errors.New("bind-err")
	}

	ifi, err := net.InterfaceByName(ifName)
	if err != nil {
		return -1, errors.New("intf-err")
	}

	ll := syscall.SockaddrLinklayer{
		Protocol: Htons(syscall.ETH_P_ARP),
		Ifindex:  ifi.Index,
		Pkttype:  0, // syscall.PACKET_HOST
		Hatype:   1,
		Halen:    6,
	}

	for i := 0; i < 6; i++ {
		ll.Addr[i] = 0xff
	}

	buf := new(bytes.Buffer)

	var sb = make([]byte, 2)
	binary.BigEndian.PutUint16(sb, 1)  // HwType = 1
	buf.Write(sb)

	binary.BigEndian.PutUint16(sb, 0x0800)  // protoType
	buf.Write(sb)

	buf.Write([]byte{6}) // hwAddrLen
	buf.Write([]byte{4}) // protoAddrLen

	binary.BigEndian.PutUint16(sb, 0x1) // OpCode
	buf.Write(sb)

	buf.Write(ifi.HardwareAddr) // senderHwAddr
	buf.Write(SIP.To4()) // senderProtoAddr

	buf.Write(bZeroAddr) // targetHwAddr
	buf.Write(DIP.To4()) // targetProtoAddr

	if err := syscall.Bind(fd, &ll); err != nil {
		return -1, errors.New("bind-err")
	}
	if err := syscall.Sendto(fd, buf.Bytes(), 0, &ll); err != nil {
		return -1, errors.New("send-err")
	}

	return 0, nil
}