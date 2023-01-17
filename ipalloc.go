// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2023 NetLOX Inc

package loxilib

import (
	"errors"
	"net"
)

// Constants
const (
	IPClusterDefault = "default"
)

// IPRange - Defines an IPRange
type IPRange struct {
	ipNet  net.IPNet
	freeID *Counter
	first  uint64
	ident  map[uint32]struct{}
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

	for i := 0; i < 8 && arrIndex >= 0 && v > 0; i++ {
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
		iplen = 4
	} else {
		iplen = 16
	}
	arrIndex := len(baseIP) - iplen
	arrIndex1 := len(IP) - iplen

	for i := 0; i < 8 && arrIndex < iplen; i++ {
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

// AllocateNewIP - Allocate a New IP address from the given cluster and CIDR range
// If id is 0, a new IP address will be allocated else IP addresses will be shared and
// it will be same as the first IP address allocted for this range
func (ipa *IPAllocator) AllocateNewIP(cluster string, cidr string, id uint32) (net.IP, error) {
	var ipCPool *IPClusterPool
	var ipr *IPRange
	var newIndex uint64
	_, ipn, err := net.ParseCIDR(cidr)

	if err != nil {
		return net.IP{0, 0, 0, 0}, errors.New("Invalid CIDR")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		if err := ipa.AddIPRange(cluster, cidr); err != nil {
			return net.IP{0, 0, 0, 0}, errors.New("No such IP Cluster Pool")
		}
		if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
			return net.IP{0, 0, 0, 0}, errors.New("IP Range allocation failure")
		}
	}

	if ipr = ipCPool.pool[cidr]; ipr == nil {
		return net.IP{0, 0, 0, 0}, errors.New("No such IP Range")
	}

	if _, ok := ipr.ident[id]; ok {
		if id != 0 {
			return net.IP{0, 0, 0, 0}, errors.New("IP Range,Ident exists")
		}
	}

	if id == 0 || ipr.first == 0 {
		newIndex, err = ipr.freeID.GetCounter()
		if err != nil {
			return net.IP{0, 0, 0, 0}, errors.New("IP Alloc counter failure")
		}
		if ipr.first == 0 {
			ipr.first = newIndex
		}
	} else {
		newIndex = ipr.first
	}

	ipr.ident[id] = struct{}{}

	retIP := addIPIndex(ipn.IP, uint64(newIndex))

	return retIP, nil
}

// DeAllocateIP - Deallocate the IP address from the given cluster and CIDR range
func (ipa *IPAllocator) DeAllocateIP(cluster string, cidr string, id uint32, IPString string) error {
	var ipCPool *IPClusterPool
	var ipr *IPRange
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return errors.New("Invalid CIDR")
	}

	IP := net.ParseIP(IPString)
	if IP == nil {
		return errors.New("Invalid IP String")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		return errors.New("IP Cluster not found")
	}

	if ipr = ipCPool.pool[cidr]; ipr == nil {
		return errors.New("No such IP Range")
	}

	if _, ok := ipr.ident[id]; !ok {
		return errors.New("IP Range - Ident not found")
	}

	retIndex := diffIPIndex(ipr.ipNet.IP, IP)
	if retIndex <= 0 {
		if retIndex != 0 || (retIndex == 0 && ipr.first != 0) {
			return errors.New("IP return index not found")
		}
	}

	delete(ipr.ident, id)

	if len(ipr.ident) == 0 {
		err = ipr.freeID.PutCounter(retIndex)
		if err != nil {
			return errors.New("IP Range counter failure")
		}
	}

	return nil
}

// AddIPRange - Add a new IP Range for allocation in a cluster
func (ipa *IPAllocator) AddIPRange(cluster string, cidr string) error {
	var ipCPool *IPClusterPool

	ip, ipn, err := net.ParseCIDR(cidr)
	if err != nil {
		return errors.New("Invalid CIDR")
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
		return errors.New("Can't find IP Cluster Pool")
	}

	for _, ipr := range ipCPool.pool {
		if ipr.ipNet.Contains(ip) {
			return errors.New("Existing IP Pool")
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
		return errors.New("IP Pool subnet error")
	}

	if iprSz > uint64(^uint16(0)) {
		iprSz = uint64(^uint16(0))
	}

	// If it is a x.x.x.0/24, then we will allocate
	// from x.x.x.1 to x.x.x.254
	ipr.freeID = NewCounter(start, iprSz)

	if ipr.freeID == nil {
		return errors.New("IP Pool alloc failed")
	}

	ipr.ident = make(map[uint32]struct{})
	ipCPool.pool[cidr] = ipr

	return nil
}

// DeleteIPRange - Delete a IP Range from allocation in a cluster
func (ipa *IPAllocator) DeleteIPRange(cluster string, cidr string) error {
	var ipCPool *IPClusterPool
	_, _, err := net.ParseCIDR(cidr)

	if err != nil {
		return errors.New("Invalid CIDR")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		return errors.New("No such IP Cluster Pool")
	}

	if ipr := ipCPool.pool[cidr]; ipr == nil {
		return errors.New("No such IP Range")
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
