// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2023 NetLOX Inc

package loxilib

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	IPZoneDefault = "default"
)

type IpRange struct {
	ipNet   net.IPNet
    freeID  *Counter
	ident   map[uint16]struct{}
}

type IpZone struct {
	name    string
	pool    map[string]*IpRange
}

type IpAllocator struct {
	ipBlocks map[string]*IpZone
}

func (ipa *IpAllocator) AddNewRange(zone string, cidr string) error {
	ip, ipn, err:= net.ParseCIDR(cidr)
	
	if err != nil {
		return errors.New("Invalid CIDR")
	}

	ipz := ipa.ipBlocks[IPZoneDefault]

	if zone != IPZoneDefault {
		if ipz = ipa.ipBlocks[zone]; ipz == nil {
			ipz := new(IpZone)
			ipa.ipBlocks[IPZoneDefault] = ipz
		}
	}

	if ipz == nil {
		return errors.New("Can't find IP Zone")
	}

	for _, ipr := range ipz.pool {
		if ipr.ipNet.Contains(ip) {
			return errors.New("Existing IP Pool")
		}
	}

}

func IpAllocatorNew() *IpAllocator {
	ipa := new(IpAllocator)
	ipa.ipBlocks = make(map[string]*IpZone)

	ipz := new(IpZone)
	ipa.ipBlocks[IPZoneDefault] = ipz

	return ipa
}