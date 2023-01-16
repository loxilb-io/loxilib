// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2023 NetLOX Inc

package loxilib

import (
	"errors"
	"net"
)

const (
	IPClusterDefault = "default"
)

type IpRange struct {
	ipNet  net.IPNet
	freeID *Counter
	ident  map[uint32]struct{}
}

type IpClusterPool struct {
	name string
	pool map[string]*IpRange
}

type IpAllocator struct {
	ipBlocks map[string]*IpClusterPool
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

func (ipa *IpAllocator) AllocateNewIP(cluster string, cidr string, id uint32) (net.IP, error) {
	var ipCPool *IpClusterPool
	var ipr *IpRange
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
		return net.IP{0, 0, 0, 0}, errors.New("IP Range,Ident exists")
	}

	newIndex, err := ipr.freeID.GetCounter()
	if err != nil {
		return net.IP{0, 0, 0, 0}, errors.New("IP Alloc counter failure")
	}

	ipr.ident[id] = struct{}{}

	retIP := addIPIndex(ipn.IP, uint64(newIndex))

	return retIP, nil
}

func (ipa *IpAllocator) DeAllocateIP(cluster string, cidr string, id uint32, IPString string) error {
	var ipCPool *IpClusterPool
	var ipr *IpRange
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
		return errors.New("IP return index not found")
	}

	err = ipr.freeID.PutCounter(retIndex)
	if err != nil {
		return errors.New("IP Range counter failure")
	}

	delete(ipr.ident, id)

	return nil
}

func (ipa *IpAllocator) AddIPRange(cluster string, cidr string) error {
	var ipCPool *IpClusterPool
	ip, ipn, err := net.ParseCIDR(cidr)

	if err != nil {
		return errors.New("Invalid CIDR")
	}

	ipCPool = ipa.ipBlocks[IPClusterDefault]

	if cluster != IPClusterDefault {
		if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
			ipCPool = new(IpClusterPool)
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

	ipr := new(IpRange)
	ipr.ipNet = *ipn
	iprSz := uint64(0)
	sz, _ := ipn.Mask.Size()
	if IsNetIPv4(ip.String()) {
		if sz == 31 {
			iprSz = 1
		} else {
			iprSz = (1 << (32 - sz)) - 2
		}
	} else {
		if sz == 127 {
			iprSz = 1
		} else {
			iprSz = (1 << (128 - sz)) - 2
		}
	}

	if iprSz < 1 {
		return errors.New("IP Pool subnet error")
	}

	if iprSz > uint64(^uint16(0)) {
		iprSz = uint64(^uint16(0))
	}

	// If it is a x.x.x.0/24, then we will allocate
	// from x.x.x.1 to x.x.x.254
	ipr.freeID = NewCounter(1, iprSz)

	if ipr.freeID == nil {
		return errors.New("IP Pool alloc failed")
	}

	ipr.ident = make(map[uint32]struct{})
	ipCPool.pool[cidr] = ipr

	return nil
}

func (ipa *IpAllocator) DeleteIPRange(cluster string, cidr string) error {
	var ipCPool *IpClusterPool
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

func IpAllocatorNew() *IpAllocator {
	ipa := new(IpAllocator)
	ipa.ipBlocks = make(map[string]*IpClusterPool)

	ipCPool := new(IpClusterPool)
	ipCPool.pool = make(map[string]*IpRange)
	ipa.ipBlocks[IPClusterDefault] = ipCPool

	return ipa
}
