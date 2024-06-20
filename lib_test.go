// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"fmt"
	"net"
	"testing"

	"github.com/loxilb-io/sctp"
)

type Tk struct {
}

func (tk *Tk) TrieNodeWalker(b string) {
	fmt.Printf("%s\n", b)
}

func (tk *Tk) TrieData2String(d TrieData) string {

	if data, ok := d.(int); ok {
		return fmt.Sprintf("%d", data)
	}

	return ""
}

var tk Tk

func BenchmarkTrie(b *testing.B) {
	var tk Tk
	trieR := TrieInit(false)

	i := 0
	j := 0
	k := 0
	pLen := 32

	for n := 0; n < b.N; n++ {
		i = n & 0xff
		j = n >> 8 & 0xff
		k = n >> 16 & 0xff

		/*if j > 0 {
		      pLen = 24
		  } else {
		      pLen = 32
		  }*/
		route := fmt.Sprintf("192.%d.%d.%d/%d", k, j, i, pLen)
		res := trieR.AddTrie(route, n)
		if res != 0 {
			b.Fatalf("failed to add %s:%d - (%d)", route, n, res)
			trieR.Trie2String(&tk)
		}
	}
}

func TestTrie(t *testing.T) {
	trieR := TrieInit(false)
	route := "192.168.1.1/32"
	data := 1100
	res := trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/15"
	data = 100
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/16"
	data = 99
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/8"
	data = 1
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/16"
	data = 1
	res = trieR.AddTrie(route, data)
	if res == 0 {
		t.Fatalf("re-added %s:%d", route, data)
	}

	route = "0.0.0.0/0"
	data = 222
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "8.8.8.8/32"
	data = 1200
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "10.10.10.10/32"
	data = 12
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "1.1.1.1/32"
	data = 1212
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "14.1.48.16/28"
	data = 120
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	// If we need to dump trie elements
	// Run # go test -v .
	trieR.Trie2String(&tk)

	ret, ipn, rdata := trieR.FindTrie("192.41.3.1")
	if ret != 0 || (*ipn).String() != "192.0.0.0/8" || rdata != 1 {
		t.Fatalf("failed to find %s", "192.41.3.1")
	}

	ret1, ipn, rdata1 := trieR.FindTrie("195.41.3.1")
	if ret1 != 0 || (*ipn).String() != "0.0.0.0/0" || rdata1 != 222 {
		t.Fatalf("failed to find %s", "195.41.3.1")
	}

	ret2, ipn, rdata2 := trieR.FindTrie("8.8.8.8")
	if ret2 != 0 || (*ipn).String() != "8.8.8.8/32" || rdata2 != 1200 {
		t.Fatalf("failed to find %d %s %d", ret, "8.8.8.8", rdata)
	}

	ret, ipn, rdata = trieR.FindTrie("14.1.48.28")
	if ret != 0 || (*ipn).String() != "14.1.48.16/28" || rdata != 120 {
		t.Fatalf("failed to find %d %s %d", ret, "14.1.48.28", rdata)
	}

	route = "0.0.0.0/0"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	ret1, _, _ = trieR.FindTrie("195.41.3.1")
	if ret1 == 0 {
		t.Fatalf("failed to find %s", "195.41.3.1")
	}

	route = "192.168.1.1/32"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	route = "192.168.1.0/15"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	route = "192.168.1.0/16"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	route = "192.168.1.0/8"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	route = "0.0.0.0/0"
	res = trieR.DelTrie(route)
	if res == 0 {
		t.Fatalf("failed to delete %s", route)
	}

	route = "8.8.8.8/32"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	route = "10.10.10.10/32"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	route = "1.1.1.1/24"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}

	trieR6 := TrieInit(true)
	route = "2001:db8::/32"
	data = 5100
	res = trieR6.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	route = "2001:db8::1/128"
	data = 5200
	res = trieR6.AddTrie(route, data)
	if res != 0 {
		t.Fatalf("failed to add %s:%d", route, data)
	}

	ret, ipn, rdata = trieR6.FindTrie("2001:db8::1")
	if ret != 0 || (*ipn).String() != "2001:db8::1/128" || rdata != 5200 {
		t.Fatalf("failed to find %s", "2001:db8::1")
	}

	route = "2001:db8::1/128"
	res = trieR6.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to del %s", route)
	}

	route = "2001:db8::/32"
	res = trieR6.DelTrie(route)
	if res != 0 {
		t.Fatalf("failed to delete %s", route)
	}
}

func TestCounter(t *testing.T) {

	cR := NewCounter(0, 10)

	for i := 0; i < 12; i++ {
		idx, err := cR.GetCounter()

		if err == nil {
			if i > 9 {
				t.Fatalf("get Counter unexpected %d:%d", i, idx)
			}
		} else {
			if i <= 9 {
				t.Fatalf("failed to get Counter %d:%d", i, idx)
			}
		}
	}

	err := cR.PutCounter(5)
	if err != nil {
		t.Fatalf("failed to put valid Counter %d", 5)
	}

	err = cR.PutCounter(2)
	if err != nil {
		t.Fatalf("failed to put valid Counter %d", 2)
	}

	err = cR.PutCounter(15)
	if err == nil {
		t.Fatalf("Able to put invalid Counter %d", 15)
	}

	err = cR.ReserveCounter(15)
	if err == nil {
		t.Fatalf("Able to put reserve Counter %d:%s", 15, err)
	}

	var idx uint64
	idx, err = cR.GetCounter()
	if idx != 5 || err != nil {
		t.Fatalf("Counter get got %d of expected %d", idx, 5)
	}

	idx, err = cR.GetCounter()
	if idx != 2 || err != nil {
		t.Fatalf("Counter get got %d of expected %d", idx, 2)
	}

	idx, err = cR.GetCounter()
	if idx != ^uint64(0) || err == nil {
		t.Fatalf("Counter get got %d", idx)
	}

	err = cR.PutCounter(2)
	if err != nil {
		t.Fatalf("failed to put valid Counter %d", 2)
	}

	err = cR.ReserveCounter(2)
	if err != nil {
		t.Fatalf("failed to reserv valid Counter %d", 2)
	}

	_, err = cR.GetCounter()
	if err == nil {
		t.Fatalf("Counter get passed unexpectedly %s", err)
	}

	err = cR.PutCounter(2)
	if err != nil {
		t.Fatalf("failed to put valid Counter %d", 2)
	}

	_, err = cR.GetCounter()
	if err != nil {
		t.Fatalf("Counter get failed unexpectedly %s", err)
	}

	cR = NewCounter(0, 5)
	err = cR.ReserveCounter(0)
	if err != nil {
		t.Fatalf("failed to reserve valid Counter %d", 0)
	}

	idx, err = cR.GetCounter()
	if err != nil {
		t.Fatalf("failed to get valid Counter %d", 0)
	}

	if idx == 0 {
		t.Fatalf("reservation failed Counter %d", 1)
	}
}

func TestIfStat(t *testing.T) {
	var ifs IfiStat

	err := NetGetIfiStats("eth0", &ifs)
	if err != 0 {
		t.Fatalf("Get stats failed for eth0")
	}
}

func TestIPAlloc(t *testing.T) {
	ipa := IpAllocatorNew()
	proto := "tcp"

	ipa.AddIPRange(IPClusterDefault, "123.123.123.0/24")
	ipa.ReserveIP(IPClusterDefault, "123.123.123.0/24", IPAMNoIdent, "123.123.123.2")
	for i := 0; i < 255; i++ {
		ip, err := ipa.AllocateNewIP(IPClusterDefault, "123.123.123.0/24", IPAMNoIdent)
		if i >= 253 && err == nil {
			t.Fatal("Failed IP Alloc for 123.123.123.0/24 - Check Alloc Algo")
		} else if i < 253 && err != nil {
			t.Fatalf("Failed IP Alloc for 123.123.123.0/24 : %d:%s", i, err)
		} else if err == nil {
			expected := ""
			if i < 1 {
				expected = fmt.Sprintf("123.123.123.%d", i+1)
			} else {
				expected = fmt.Sprintf("123.123.123.%d", i+2)
			}
			if ip.String() != expected {
				t.Fatalf("Failed IP Alloc for 123.123.123.0/24: %s:%s", ip.String(), expected)
			}
		}
	}

	err := ipa.DeAllocateIP(IPClusterDefault, "123.123.123.0/24", IPAMNoIdent, "123.123.123.1")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "123.123.123.1", err)
	}

	ip, err := ipa.AllocateNewIP(IPClusterDefault, "123.123.123.0/24", IPAMNoIdent)
	if err != nil || ip.String() != "123.123.123.1" {
		t.Fatalf("Failed IP Alloc for 123.123.123.0/24:%s", "123.123.123.1")
	}

	ipa.AddIPRange(IPClusterDefault, "11.11.11.0/31")
	ip, err = ipa.AllocateNewIP(IPClusterDefault, "11.11.11.0/31", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 11.11.11.0/31 - Check Alloc Algo")
	}

	if ip.String() != "11.11.11.0" {
		t.Fatalf("Failed IP Alloc for 11.11.11.0/31: %s:%s", ip.String(), "11.11.11.0")
	}

	ip, err = ipa.AllocateNewIP(IPClusterDefault, "11.11.11.0/31", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 11.11.11.0/31 - Check Alloc Algo")
	}

	if ip.String() != "11.11.11.1" {
		t.Fatalf("Failed IP Alloc for 11.11.11.0/31: %s:%s", ip.String(), "11.11.11.1")
	}

	_, err = ipa.AllocateNewIP(IPClusterDefault, "11.11.11.0/31", IPAMNoIdent)
	if err == nil {
		t.Fatal("Invalid IP Alloc for 11.11.11.0/31 - Check Alloc Algo")
	}

	err = ipa.DeleteIPRange(IPClusterDefault, "11.11.11.0/31")
	if err != nil {
		t.Fatal("Failed to delete IP Alloc for 11.11.11.0/31 - Check Alloc Algo")
	}

	err = ipa.AddIPRange(IPClusterDefault, "12.12.0.0/16")
	if err != nil {
		t.Fatal("Failed to Add IP Range for 12.12.0.0/16")
	}

	ip1, err := ipa.AllocateNewIP(IPClusterDefault, "12.12.0.0/16", IPAMNoIdent)
	if err != nil {
		t.Fatalf("IP Alloc failed for 12.12.0.0/16:1:%s", err)
	}

	ip2, err := ipa.AllocateNewIP(IPClusterDefault, "12.12.0.0/16", MakeIPAMIdent("", 1, proto))
	if err != nil {
		t.Fatalf("IP Alloc failed for 12.12.0.0/16:2:%s", err)
	}
	if ip2.String() != "12.12.0.1" {
		t.Fatalf("Shared IP Alloc failed for 12.12.0.0/16:2:%s", err)
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "12.12.0.0/16", IPAMNoIdent, ip1.String())
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", ip1.String(), err)
	}

	_, err = ipa.AllocateNewIP(IPClusterDefault, "12.12.0.0/16", IPAMNoIdent)
	if err != nil {
		t.Fatalf("IP Alloc failed for 12.12.0.0/16:1:%s", err)
	}

	err = ipa.AddIPRange(IPClusterDefault, "3ffe::/64")
	if err != nil {
		t.Fatal("Failed to Add IP Range for 3ffe::/64")
	}

	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "3ffe::/64", IPAMNoIdent)
	if err != nil {
		t.Fatalf("IP Alloc failed for 3ffe::/64:%s", err)
	}

	if ip1.String() != "3ffe::1" {
		t.Fatalf("IP Alloc failed - 3ffe::1:%s", ip1.String())
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "3ffe::/64", IPAMNoIdent, ip1.String())
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", ip1.String(), err)
	}

	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "3ffe::/64", IPAMNoIdent)
	if err != nil {
		t.Fatalf("IP Alloc failed for 3ffe::/64:%s", err)
	}

	if ip1.String() != "3ffe::2" {
		t.Fatalf("IP Alloc failed - 3ffe::1:%s pass2", ip1.String())
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "3ffe::/64", IPAMNoIdent, ip1.String())
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s pass2", ip1.String(), err)
	}

	err = ipa.DeleteIPRange(IPClusterDefault, "3ffe::/64")
	if err != nil {
		t.Fatal("Failed to delete IP Alloc for 3ffe::/64 - Check Alloc Algo")
	}

	_, err = ipa.AllocateNewIP(IPClusterDefault, "3ffe::/64", IPAMNoIdent)
	if err == nil {
		t.Fatalf("IP Alloc unexpected success for 3ffe::/64:%s", err)
	}

	err = ipa.AddIPRange(IPClusterDefault, "4ffe::/64")
	if err != nil {
		t.Fatal("Failed to Add IP Range for 4ffe::/64")
	}

	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "4ffe::/64", MakeIPAMIdent("", 1, proto))
	if err != nil {
		t.Fatalf("IP Alloc failed for 4ffe::/64:%s", err)
	}

	if ip1.String() != "4ffe::1" {
		t.Fatalf("IP Alloc failed - 4ffe::1:%s", ip1.String())
	}

	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "4ffe::/64", MakeIPAMIdent("", 2, proto))
	if err != nil {
		t.Fatalf("IP Alloc failed for 4ffe::/64:%s", err)
	}

	if ip1.String() != "4ffe::1" {
		t.Fatalf("IP Alloc failed - 4ffe::1:%s", ip1.String())
	}

	err = ipa.DeleteIPRange(IPClusterDefault, "4ffe::/64")
	if err != nil {
		t.Fatal("Failed to delete IP Alloc for 4ffe::/64 - Check Alloc Algo")
	}

	err = ipa.AddIPRange(IPClusterDefault, "100.100.100.1/32")
	if err != nil {
		t.Fatal("Failed to Add IP Range for 100.100.100.1/32")
	}

	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "100.100.100.1/32", IPAMNoIdent)
	if err != nil {
		t.Fatalf("IP Alloc failed for 100.100.100.1/32:%s", err)
	}

	if ip1.String() != "100.100.100.1" {
		t.Fatalf("IP Alloc failed - 100.100.100.1:%s", ip1.String())
	}

	_, err = ipa.AllocateNewIP(IPClusterDefault, "100.100.100.1/32", IPAMNoIdent)
	if err == nil {
		t.Fatalf("IP Alloc should fail for 100.100.100.1/32:%s", err)
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "100.100.100.1/32", IPAMNoIdent, "100.100.100.1")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "100.100.100.1", err)
	}

	_, err = ipa.AllocateNewIP(IPClusterDefault, "100.100.100.1/32", IPAMNoIdent)
	if err != nil {
		t.Fatalf("IP Alloc failed for 100.100.100.1/32:%s", err)
	}

	err = ipa.AddIPRange(IPClusterDefault, "74.125.227.24/29")
	if err != nil {
		t.Fatalf("IP Alloc failed for 74.125.227.24/29:%s", err)
	}

	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "74.125.227.24/29", IPAMNoIdent)
	if err != nil {
		t.Fatalf("IP Alloc failed for 100.100.100.1/32:%s", err)
	}

	if ip1.String() != "74.125.227.24" {
		t.Fatalf("IP Alloc failed for 74.125.227.24:%s", ip1.String())
	}

	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "74.125.227.24/29", MakeIPAMIdent("", 1, proto))
	if err != nil {
		t.Fatalf("IP Alloc failed for 100.100.100.1/32:%s", err)
	}

	if ip1.String() != "74.125.227.24" {
		t.Fatalf("IP Alloc failed for 74.125.227.24:%s", ip1.String())
	}

	err = ipa.DeleteIPRange(IPClusterDefault, "74.125.227.24/29")
	if err != nil {
		t.Fatalf("IP Delete Range failed for 74.125.227.24/29:%s", err)
	}

	ipa.AddIPRange(IPClusterDefault, "71.71.71.0/31")
	err = ipa.ReserveIP(IPClusterDefault, "71.71.71.0/31", IPAMNoIdent, "71.71.71.0")
	if err != nil {
		t.Fatal("Failed to reserve IP 71.71.71.0 - Check Alloc Algo")
	}

	ip, err = ipa.AllocateNewIP(IPClusterDefault, "71.71.71.0/31", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 71.71.71.0/31 - Check Alloc Algo")
	}

	if ip.String() != "71.71.71.1" {
		t.Fatalf("Failed IP Alloc for 71.71.71.0/31: %s:%s", ip.String(), "71.71.71.1")
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "71.71.71.0/31", IPAMNoIdent, "71.71.71.0")
	if err != nil {
		t.Fatalf("Failed IP DeAlloc for 71.71.71.0/31:%s:%s", "71.71.71.0", err)
	}

	ip, err = ipa.AllocateNewIP(IPClusterDefault, "71.71.71.0/31", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 71.71.71.0/31 - Check Alloc Algo")
	}

	if ip.String() != "71.71.71.0" {
		t.Fatalf("Failed IP Alloc for 71.71.71.0/31: %s:%s", ip.String(), "71.71.71.0")
	}

	ipa.AddIPRange(IPClusterDefault, "192.168.10.252/24")
	ip, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.10.252/24", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 192.168.10.252/24 - Check Alloc Algo")
	}

	if ip.String() != "192.168.10.252" {
		t.Fatalf("Failed IP Alloc for 192.168.10.252/24: %s:%s", ip.String(), "192.168.10.252")
	}

	ip, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.10.252/24", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 192.168.10.252/24 - Check Alloc Algo")
	}

	if ip.String() != "192.168.10.253" {
		t.Fatalf("Failed IP Alloc for 192.168.10.252/24: %s:%s", ip.String(), "192.168.10.253")
	}

	ip, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.10.252/24", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 192.168.10.252/24 - Check Alloc Algo")
	}

	if ip.String() != "192.168.10.254" {
		t.Fatalf("Failed IP Alloc for 192.168.10.252/24: %s:%s", ip.String(), "192.168.10.254")
	}

	_, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.10.252/24", IPAMNoIdent)
	if err == nil {
		t.Fatal("Allocated IP Alloc for 192.168.10.252/24 - Check Alloc Algo")
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.10.252/24", IPAMNoIdent, "192.168.10.253")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "192.168.10.253", err)
	}

	ip, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.10.252/24", IPAMNoIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 192.168.10.252/24 - Check Alloc Algo")
	}

	if ip.String() != "192.168.10.253" {
		t.Fatalf("Failed IP Alloc for 192.168.10.252/24: %s:%s", ip.String(), "192.168.10.253")
	}

	err = ipa.AddIPRange(IPClusterDefault, "192.168.82.210/29")
	if err != nil {
		t.Fatal("Failed to Add IP Range for 100.100.100.1/32")
	}

	ipamIdent := MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29%s", err)
		}
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err == nil {
			t.Fatalf("IP unexpectedly alloced for 192.168.82.210/29:%s", err)
		}
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.82.210/29", IPAMNoIdent, "192.168.82.212")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "192.168.82.211", err)
	}

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	if ip1.String() != "192.168.82.212" {
		t.Fatalf("IP alloced is not 192.168.82.211 for 192.168.82.210/29:%s", err)
	}

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.82.210/29", ipamIdent, "192.168.82.210")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "192.168.82.210", err)
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.82.210/29", IPAMNoIdent, "192.168.82.212")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "192.168.82.212", err)
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.82.210/29", IPAMNoIdent, "192.168.82.213")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "192.168.82.213", err)
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.82.210/29", IPAMNoIdent, "192.168.82.214")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "192.168.82.214", err)
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.82.210/29", IPAMNoIdent, "192.168.82.214")
	if err == nil {
		t.Fatalf("IP DeAlloc success unexpected for %s:%s", "192.168.82.214", err)
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "192.168.82.210/29", IPAMNoIdent, "192.168.82.211")
	if err != nil {
		t.Fatalf("IP DeAlloc failed for %s:%s", "192.168.82.211", err)
	}

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	if ip1.String() != "192.168.82.210" {
		t.Fatalf("IP alloced is not 192.168.82.210 for 192.168.82.210/29:%s", err)
	}

	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	if ip1.String() == "192.168.82.210" {
		t.Fatalf("IP alloced is  %s not 192.168.82.211 for 192.168.82.210/29:%s", ip1.String(), err)
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	if ip1.String() == "192.168.82.210" {
		t.Fatalf("IP alloced is not 192.168.82.212 for 192.168.82.210/29:%s", err)
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	if ip1.String() == "192.168.82.210" {
		t.Fatalf("IP alloced is not 192.168.82.212 for 192.168.82.210/29:%s", err)
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
	if err != nil {
		ipamIdent = IPAMNoIdent
		ip1, err = ipa.AllocateNewIP(IPClusterDefault, "192.168.82.210/29", ipamIdent)
		if err != nil {
			t.Fatalf("IP Alloc failed for 192.168.82.210/29:%s", err)
		}
	}

	if ip1.String() == "192.168.82.210" {
		t.Fatalf("IP alloced is not 192.168.82.212 for 192.168.82.210/29:%s", err)
	}
	fmt.Printf("allocIP %s\n", ip1.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ipa.AddIPRange(IPClusterDefault, "0.0.0.0/32")
	ip0, err := ipa.AllocateNewIP(IPClusterDefault, "0.0.0.0/32", ipamIdent)
	if err != nil {
		t.Fatal("Failed IP Alloc for 0.0.0.0 - Check Alloc Algo")
	}

	fmt.Printf("allocIP %s\n", ip0.String())

	ipamIdent = MakeIPAMIdent("", 80, "tcp")
	ipa.AddIPRange(IPClusterDefault, "0.0.0.0/32")
	ip, err = ipa.AllocateNewIP(IPClusterDefault, "0.0.0.0/32", ipamIdent)
	if err == nil {
		t.Fatal("Failed IP Alloc for 0.0.0.0 - Check Alloc Algo")
	} else {
		fmt.Printf("Hello --- %s\n", err)
	}

	ipamIdent1 := MakeIPAMIdent("", 90, "tcp")
	ip, err = ipa.AllocateNewIP(IPClusterDefault, "0.0.0.0/32", ipamIdent1)
	if err != nil {
		t.Fatal("Failed IP Alloc for 0.0.0.0 - Check Alloc Algo")
	} else {
		fmt.Printf("Hello ---  alloc %s\n", ip.String())
	}

	err = ipa.DeAllocateIP(IPClusterDefault, "0.0.0.0/32", ipamIdent, ip0.String())
	if err != nil {
		t.Fatal("Failed IP DeAlloc for 0.0.0.0 - Check Alloc Algo")
	}
}

func TestProber(t *testing.T) {
	epIp, err := net.ResolveIPAddr("ip", "127.0.0.1")
	if err != nil {
		t.Fatal("Failed to resolve IP")
	}

	ips := []net.IPAddr{*epIp}

	addr := &sctp.SCTPAddr{
		IPAddrs: ips,
		Port:    8080,
	}

	cn, err := sctp.DialSCTP("sctp", nil, addr, false)
	if err == nil {
		t.Logf("sctp prober connected")
		cn.Close()
	} else {
		t.Logf("sctp prober not connected %s", err)
	}

	cn, err = sctp.DialSCTP("sctp", nil, addr, true)
	if err == nil {
		t.Logf("sctp (block) prober connected")
		cn.Close()
	} else {
		t.Logf("sctp (block) prober not connected %s", err)
	}

	// We are checking if this blocks forever
	t.Logf("sctp prober test OK")

	sOk := L4ServiceProber("udp", "192.168.20.55:12234", "", "", "")
	t.Logf("udp prober test1 %v\n", sOk)

	sOk = L4ServiceProber("udp", "127.0.0.1:12234", "", "", "")
	t.Logf("udp prober test2 %v\n", sOk)
}
