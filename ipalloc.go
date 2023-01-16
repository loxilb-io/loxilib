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
	ident  map[uint16]struct{}
}

type IpClusterPool struct {
	name string
	pool map[string]*IpRange
}

type IpAllocator struct {
	ipBlocks map[string]*IpClusterPool
}

func (ipa *IpAllocator) AllocateNewIP(cluster string, cidr string) (net.IP, error) {
	var ipCPool *IpClusterPool
	_, _, err := net.ParseCIDR(cidr)

	if err != nil {
		return net.IP{0, 0, 0, 0}, errors.New("Invalid CIDR")
	}

	if ipCPool = ipa.ipBlocks[cluster]; ipCPool == nil {
		return net.IP{0, 0, 0, 0}, errors.New("No such IP Cluster Pool")
	}

	if ipr := ipCPool.pool[cidr]; ipr == nil {
		return net.IP{0, 0, 0, 0}, errors.New("No such IP Range")
	}

	return net.IP{0, 0, 0, 0}, nil
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
	iprSz := 0
	sz, _ := ipn.Mask.Size()
	if IsNetIPv4(ip.String()) {
		iprSz = (1 << (32 - sz)) - 1
	} else {
		iprSz = (1 << (128 - sz)) - 1
	}

	// If it is a x.x.x.0/8, then we will allocate
	// from x.x.x.1 to x.x.x.255
	ipr.freeID = NewCounter(1, iprSz)

	if ipr.freeID == nil {
		return errors.New("IP Pool alloc failed")
	}

	ipr.ident = make(map[uint16]struct{})
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

	ipz := new(IpClusterPool)
	ipa.ipBlocks[IPClusterDefault] = ipz

	return ipa
}
