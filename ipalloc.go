// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2023 NetLOX Inc

package loxilib

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Constants
const (
	IPClusterDefault = "default"
	IP4Len           = 4
	IP6Len           = 16
	IPAMNoIdent      = ""
)

// IdentKey - key of IP Pool
type IdentKey string

// Generate a key with a combination of id and port.
func getIdentKey(idString string) IdentKey {
	lowerProto := strings.ToLower(idString)
	return IdentKey(lowerProto)
}

// Make a identifier with a combination of name, id and port.
func MakeIPAMIdent(name string, id uint32, proto string) string {
	lowerProto := strings.ToLower(string(proto))
	return fmt.Sprintf("%s|%d|%s", name, id, lowerProto)
}

// IPRange - Defines an IPRange
type IPRange struct {
	ipNet  net.IPNet
	freeID *Counter
	fOK    bool
	first  uint64
	ident  map[IdentKey]struct{}
}

// IPClusterPool - Holds IP ranges for a cluster
type IPClusterPool struct {
	name string
	pool map[string]*IPRange
}

// IPAllocator - Main IP allocator context
type IPAllocator struct {
	ipBlocks map[string]*IPClusterPool
}

func addIPIndex(ip net.IP, index uint64) net.IP {
	retIP := ip

	v := index
	c := uint64(0)
	arrIndex := len(ip) - 1

	for i := 0; i < IP6Len && arrIndex >= 0 && v > 0; i++ {
		c = v / 255
		retIP[arrIndex] += uint8((v + c) % 255)
		arrIndex--
		v >>= 8
	}

	return retIP
}

func diffIPIndex(baseIP net.IP, IP net.IP) uint64 {
	var index uint64
	iplen := 0
	if IsNetIPv4(baseIP.String()) {
		iplen = IP4Len
	} else {
		iplen = IP6Len
	}
	arrIndex := len(baseIP) - iplen
	arrIndex1 := len(IP) - iplen

	for i := 0; i < IP6Len && arrIndex < iplen; i++ {
		basev := uint8(baseIP[arrIndex])
		ipv := uint8(IP[arrIndex1])
		if basev > ipv {
			return 0
		}

		index = uint64(ipv - basev)
		arrIndex++
		arrIndex1++
		index |= index << (8 * (iplen - i - 1))
	}

	return index
}

// ReserveIP - Don't allocate this IP address/ID pair from the given cluster and CIDR range
// If id is empty, a new IP address will be allocated else IP addresses will be shared and
// it will be same as the first IP address allocted for this range
func (ipa *IPAllocator) ReserveIP(cluster string, cidr string, idString string, IPString string) error {
	var ipCPool *IPClusterPool
	var ipr *IPRange
	_, ipn, err := net.ParseCIDR(cidr)

	if err != nil {
		return errors.New("invalid CIDR")
	}

	IP := net.ParseIP(IPString)
	if IP == nil {
		return errors.New("invalid IP String")
	}

	if !ipn.Contains(IP) {
		return errors.New("ip string out of bounds")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		if err := ipa.AddIPRange(cluster, cidr); err != nil {
			return errors.New("no such IP Cluster Pool")
		}
		if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
			return errors.New("ip range allocation failure")
		}
	}

	if ipr = ipCPool.pool[cidr]; ipr == nil {
		return errors.New("no such IP Range")
	}

	key := getIdentKey(idString)
	if _, ok := ipr.ident[key]; ok {
		if idString != "" {
			return errors.New("ip Range,Ident,proto exists")
		}
	}

	if idString == "" || !ipr.fOK {
		retIndex := diffIPIndex(ipr.ipNet.IP, IP)
		if retIndex <= 0 {
			if retIndex != 0 || (retIndex == 0 && ipr.first != 0) {
				return errors.New("ip return index not found")
			}
		}
		err = ipr.freeID.ReserveCounter(retIndex)
		if err != nil {
			return errors.New("ip reserve counter failure")
		}
		if !ipr.fOK {
			ipr.first = retIndex
			ipr.fOK = true
		}
	}

	ipr.ident[key] = struct{}{}

	return nil
}

// AllocateNewIP - Allocate a New IP address from the given cluster and CIDR range
// If idString is empty, a new IP address will be allocated else IP addresses will be shared and
// it will be same as the first IP address allocted for this range
func (ipa *IPAllocator) AllocateNewIP(cluster string, cidr string, idString string) (net.IP, error) {
	var ipCPool *IPClusterPool
	var ipr *IPRange
	var newIndex uint64
	_, ipn, err := net.ParseCIDR(cidr)

	if err != nil {
		return net.IP{0, 0, 0, 0}, errors.New("invalid CIDR")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		if err := ipa.AddIPRange(cluster, cidr); err != nil {
			return net.IP{0, 0, 0, 0}, errors.New("no such IP Cluster Pool")
		}
		if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
			return net.IP{0, 0, 0, 0}, errors.New("ip Range allocation failure")
		}
	}

	if ipr = ipCPool.pool[cidr]; ipr == nil {
		return net.IP{0, 0, 0, 0}, errors.New("no such IP Range")
	}

	key := getIdentKey(idString)
	if _, ok := ipr.ident[key]; ok {
		if idString != "" {
			return net.IP{0, 0, 0, 0}, errors.New("ip/ident exists")
		}
	}

	if idString == "" || !ipr.fOK {
		newIndex, err = ipr.freeID.GetCounter()
		if err != nil {
			return net.IP{0, 0, 0, 0}, errors.New("ip Alloc counter failure")
		}
		if !ipr.fOK {
			ipr.first = newIndex
			ipr.fOK = true
		}
	} else {
		newIndex = ipr.first
	}

	ipr.ident[key] = struct{}{}

	retIP := addIPIndex(ipn.IP, uint64(newIndex))

	return retIP, nil
}

// DeAllocateIP - Deallocate the IP address from the given cluster and CIDR range
func (ipa *IPAllocator) DeAllocateIP(cluster string, cidr string, idString, IPString string) error {
	var ipCPool *IPClusterPool
	var ipr *IPRange
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return errors.New("invalid CIDR")
	}

	IP := net.ParseIP(IPString)
	if IP == nil {
		return errors.New("invalid IP String")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		return errors.New("ip Cluster not found")
	}

	if ipr = ipCPool.pool[cidr]; ipr == nil {
		return errors.New("no such IP Range")
	}

	key := getIdentKey(idString)
	if _, ok := ipr.ident[key]; !ok {
		return errors.New("ip Range - Ident not found")
	}

	retIndex := diffIPIndex(ipr.ipNet.IP, IP)
	if retIndex <= 0 {
		if retIndex != 0 || (retIndex == 0 && ipr.first != 0) {
			return errors.New("ip return index not found")
		}
	}

	delete(ipr.ident, key)

	if len(ipr.ident) == 0 {
		err = ipr.freeID.PutCounter(retIndex)
		if err != nil {
			return errors.New("ip Range counter failure")
		}
	}

	return nil
}

// AddIPRange - Add a new IP Range for allocation in a cluster
func (ipa *IPAllocator) AddIPRange(cluster string, cidr string) error {
	var ipCPool *IPClusterPool

	ip, ipn, err := net.ParseCIDR(cidr)
	if err != nil {
		return errors.New("invalid CIDR")
	}

	ipCPool = ipa.ipBlocks[IPClusterDefault]

	if cluster != IPClusterDefault {
		if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
			ipCPool = new(IPClusterPool)
			ipCPool.name = cluster
			ipa.ipBlocks[cluster] = ipCPool
		}
	}

	if ipCPool == nil {
		return errors.New("can't find IP Cluster Pool")
	}

	for _, ipr := range ipCPool.pool {
		if ipr.ipNet.Contains(ip) {
			return errors.New("existing IP Pool")
		}
	}

	ipr := new(IPRange)
	ipr.ipNet = *ipn
	iprSz := uint64(0)
	sz, _ := ipn.Mask.Size()
	start := uint64(1)
	if IsNetIPv4(ip.String()) {
		ignore := uint64(0)
		if sz != 32 && sz%8 == 0 {
			ignore = 2
		} else {
			start = 0
		}

		val := Ntohl(IPtonl(ip))
		msk := uint32(((1 << (32 - sz)) - 1))
		if val&msk != 0 {
			start = uint64(val & msk)
			ignore = start + 1
		}

		iprSz = (1 << (32 - sz)) - ignore
	} else {
		ignore := uint64(0)
		if sz != 128 && sz%8 == 0 {
			ignore = 2
		} else {
			start = 0
		}
		iprSz = (1 << (128 - sz)) - ignore
	}

	if iprSz < 1 {
		return errors.New("ip pool subnet error")
	}

	if iprSz > uint64(^uint16(0)) {
		iprSz = uint64(^uint16(0))
	}

	// If it is a x.x.x.0/24, then we will allocate
	// from x.x.x.1 to x.x.x.254
	ipr.freeID = NewCounter(start, iprSz)

	if ipr.freeID == nil {
		return errors.New("ip pool alloc failed")
	}

	ipr.ident = make(map[IdentKey]struct{})
	ipCPool.pool[cidr] = ipr

	return nil
}

// DeleteIPRange - Delete a IP Range from allocation in a cluster
func (ipa *IPAllocator) DeleteIPRange(cluster string, cidr string) error {
	var ipCPool *IPClusterPool
	_, _, err := net.ParseCIDR(cidr)

	if err != nil {
		return errors.New("invalid CIDR")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		return errors.New("no such IP Cluster Pool")
	}

	if ipr := ipCPool.pool[cidr]; ipr == nil {
		return errors.New("no such IP Range")
	}

	delete(ipCPool.pool, cidr)

	return nil
}

// IpAllocatorNew - Create a new allocator
func IpAllocatorNew() *IPAllocator {
	ipa := new(IPAllocator)
	ipa.ipBlocks = make(map[string]*IPClusterPool)

	ipCPool := new(IPClusterPool)
	ipCPool.pool = make(map[string]*IPRange)
	ipa.ipBlocks[IPClusterDefault] = ipCPool

	return ipa
}
