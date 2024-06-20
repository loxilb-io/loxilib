// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

// return codes
const (
	TrieSuccess = -iota
	TrieErrGeneric
	TrieErrExists
	TrieErrNoEnt
	TrieErrNoMem
	TrieErrUnknown
	TrieErrPrefix
)

// constants
const (
	TrieJmpLength   = 8
	PrefixArrLenfth = (1 << (TrieJmpLength + 1)) - 1
	PrefixArrNbits  = ((PrefixArrLenfth + TrieJmpLength) & ^TrieJmpLength) / TrieJmpLength
	PtrArrLength    = (1 << TrieJmpLength)
	PtrArrNBits     = ((PtrArrLength + TrieJmpLength) & ^TrieJmpLength) / TrieJmpLength
)

// TrieData - Any user data to be associated with a trie node
type TrieData interface {
	// Empty Interface
}

// TrieIterIntf - Interface implementation needed for trie users to
// traverse and convert data
type TrieIterIntf interface {
	TrieNodeWalker(b string)
	TrieData2String(d TrieData) string
}

type trieVar struct {
	prefix [16]byte
}

type trieState struct {
	trieData        TrieData
	lastMatchLevel  int
	lastMatchPfxLen int
	lastMatchEmpty  bool
	lastMatchTv     trieVar
	matchFound      bool
	maxLevels       int
	errCode         int
}

// TrieRoot - root of a trie data structure
type TrieRoot struct {
	v6         bool
	prefixArr  [PrefixArrNbits]uint8
	ptrArr     [PtrArrNBits]uint8
	prefixData [PrefixArrLenfth]TrieData
	ptrData    [PtrArrLength]*TrieRoot
}

// TrieInit - Initialize a trie root
func TrieInit(v6 bool) *TrieRoot {
	var root = new(TrieRoot)
	root.v6 = v6
	return root
}

func prefix2TrieVar(ipPrefix net.IP, pIndex int) trieVar {
	var tv trieVar

	if ipPrefix.To4() != nil {
		ipAddr := binary.BigEndian.Uint32(ipPrefix.To4())
		tv.prefix[0] = (uint8((ipAddr >> 24)) & 0xff)
		tv.prefix[1] = (uint8((ipAddr >> 16)) & 0xff)
		tv.prefix[2] = (uint8((ipAddr >> 8)) & 0xff)
		tv.prefix[3] = (uint8(ipAddr) & 0xff)
	} else {
		for i := 0; i < 16; i++ {
			tv.prefix[i] = uint8(ipPrefix[i])
		}
	}

	return tv
}

func grabByte(tv *trieVar, pIndex int) (uint8, error) {

	if pIndex > 15 {
		return 0xff, errors.New("Out of range")
	}

	return tv.prefix[pIndex], nil
}

func cidr2TrieVar(cidr string, tv *trieVar) (pfxLen int) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return -1
	}

	pfx := ipNet.IP.Mask(ipNet.Mask)
	pfxLen, _ = ipNet.Mask.Size()
	*tv = prefix2TrieVar(pfx, 0)
	return pfxLen
}

func shrinkPrefixArrDat(arr []TrieData, startPos int) {
	if startPos < 0 || startPos >= len(arr) {
		return
	}

	for i := startPos; i < len(arr)-1; i++ {
		arr[i] = arr[i+1]
	}
}

func shrinkPtrArrDat(arr []*TrieRoot, startPos int) {
	if startPos < 0 || startPos >= len(arr) {
		return
	}

	for i := startPos; i < len(arr)-1; i++ {
		arr[i] = arr[i+1]
	}
}

func expPrefixArrDat(arr []TrieData, startPos int) {
	if startPos < 0 || startPos >= len(arr) {
		return
	}

	for i := len(arr) - 2; i >= startPos; i-- {
		arr[i+1] = arr[i]
	}
}

func expPtrArrDat(arr []*TrieRoot, startPos int) {
	if startPos < 0 || startPos >= len(arr) {
		return
	}

	for i := len(arr) - 2; i >= startPos; i-- {
		arr[i+1] = arr[i]
	}
}

func (t *TrieRoot) addTrieInt(tv *trieVar, currLevel int, rPfxLen int, ts *trieState) int {

	if rPfxLen < 0 || ts.errCode != 0 {
		return -1
	}

	// This assumes stride of length 8
	var cval uint8 = tv.prefix[currLevel]
	var nextRoot *TrieRoot

	if rPfxLen > TrieJmpLength {
		rPfxLen -= TrieJmpLength
		ptrIdx := CountSetBitsInArr(t.ptrArr[:], int(cval)-1)
		if IsBitSetInArr(t.ptrArr[:], int(cval)) == true {
			nextRoot = t.ptrData[ptrIdx]
			if nextRoot == nil {
				ts.errCode = TrieErrUnknown
				return -1
			}
		} else {
			// If no pointer exists, then allocate it
			// Make pointer references
			nextRoot = new(TrieRoot)
			if t.ptrData[ptrIdx] != nil {
				expPtrArrDat(t.ptrData[:], ptrIdx)
				t.ptrData[ptrIdx] = nil
			}
			t.ptrData[ptrIdx] = nextRoot
			SetBitInArr(t.ptrArr[:], int(cval))
		}
		return nextRoot.addTrieInt(tv, currLevel+1, rPfxLen, ts)
	} else {
		shftBits := TrieJmpLength - rPfxLen
		basePos := (1 << rPfxLen) - 1
		// Find value relevant to currently remaining prefix len
		cval = cval >> shftBits
		idx := basePos + int(cval)
		if IsBitSetInArr(t.prefixArr[:], idx) == true {
			return TrieErrExists
		}
		pfxIdx := CountSetBitsInArr(t.prefixArr[:], idx)
		if t.prefixData[pfxIdx] != 0 {
			expPrefixArrDat(t.prefixData[:], pfxIdx)
			t.prefixData[pfxIdx] = 0
		}
		SetBitInArr(t.prefixArr[:], idx)
		t.prefixData[pfxIdx] = ts.trieData
		return 0
	}
}

func (t *TrieRoot) deleteTrieInt(tv *trieVar, currLevel int, rPfxLen int, ts *trieState) int {

	if rPfxLen < 0 || ts.errCode != 0 {
		return -1
	}

	// This assumes stride of length 8
	var cval uint8 = tv.prefix[currLevel]
	var nextRoot *TrieRoot

	if rPfxLen > TrieJmpLength {
		rPfxLen -= TrieJmpLength
		ptrIdx := CountSetBitsInArr(t.ptrArr[:], int(cval)-1)
		if IsBitSetInArr(t.ptrArr[:], int(cval)) == false {
			ts.matchFound = false
			return -1
		}

		nextRoot = t.ptrData[ptrIdx]
		if nextRoot == nil {
			ts.matchFound = false
			ts.errCode = TrieErrUnknown
			return -1
		}
		nextRoot.deleteTrieInt(tv, currLevel+1, rPfxLen, ts)
		if ts.matchFound == true && ts.lastMatchEmpty == true {
			t.ptrData[ptrIdx] = nil
			shrinkPtrArrDat(t.ptrData[:], ptrIdx)
			UnSetBitInArr(t.ptrArr[:], int(cval))
		}
		if ts.lastMatchEmpty == true {
			if CountAllSetBitsInArr(t.prefixArr[:]) == 0 &&
				CountAllSetBitsInArr(t.ptrArr[:]) == 0 {
				ts.lastMatchEmpty = true
			} else {
				ts.lastMatchEmpty = false
			}
		}
		if ts.errCode != 0 {
			return -1
		}
		return 0
	} else {
		shftBits := TrieJmpLength - rPfxLen
		basePos := (1 << rPfxLen) - 1

		// Find value relevant to currently remaining prefix len
		cval = cval >> shftBits
		idx := basePos + int(cval)
		if IsBitSetInArr(t.prefixArr[:], idx) == false {
			ts.matchFound = false
			return TrieErrNoEnt
		}
		pfxIdx := CountSetBitsInArr(t.prefixArr[:], idx-1)
		// Note - This assumes that prefix data should be non-zero
		if t.prefixData[pfxIdx] != 0 {
			t.prefixData[pfxIdx] = 0
			shrinkPrefixArrDat(t.prefixData[:], pfxIdx)
			UnSetBitInArr(t.prefixArr[:], idx)
			ts.matchFound = true
			if CountAllSetBitsInArr(t.prefixArr[:]) == 0 &&
				CountAllSetBitsInArr(t.ptrArr[:]) == 0 {
				ts.lastMatchEmpty = true
			}

			return 0
		}
		ts.matchFound = false
		ts.errCode = TrieErrUnknown
		return -1
	}
}

func (t *TrieRoot) findTrieInt(tv *trieVar, currLevel int, ts *trieState) int {

	var idx int = 0
	if ts.errCode != 0 {
		return -1
	}

	// This assumes stride of length 8
	var cval uint8 = tv.prefix[currLevel]
	ocval := cval

	for rPfxLen := TrieJmpLength; rPfxLen >= 0; rPfxLen-- {
		shftBits := TrieJmpLength - rPfxLen
		basePos := (1 << rPfxLen) - 1
		// Find value relevant to currently remaining prefix len
		cval = ocval >> shftBits
		idx = basePos + int(cval)
		pfxVal := (idx - basePos) << shftBits

		if IsBitSetInArr(t.prefixArr[:], idx) {
			ts.lastMatchLevel = currLevel
			ts.lastMatchPfxLen = 8*currLevel + rPfxLen
			ts.matchFound = true
			ts.lastMatchTv.prefix[currLevel] = byte(pfxVal)
			break
		}
	}

	cval = tv.prefix[currLevel]
	ptrIdx := CountSetBitsInArr(t.ptrArr[:], int(cval)-1)
	if IsBitSetInArr(t.ptrArr[:], int(cval)) {
		if t.ptrData[ptrIdx] != nil {
			nextRoot := t.ptrData[ptrIdx]
			ts.lastMatchTv.prefix[currLevel] = byte(cval)
			nextRoot.findTrieInt(tv, currLevel+1, ts)
		}
	}

	if ts.lastMatchLevel == currLevel {
		pfxIdx := CountSetBitsInArr(t.prefixArr[:], idx-1)
		ts.trieData = t.prefixData[pfxIdx]
	}

	return 0
}

func (t *TrieRoot) walkTrieInt(tv *trieVar, level int, ts *trieState, tf TrieIterIntf) int {
	var p int
	var pfxIdx int
	var pfxStr string

	n := 1
	pfxLen := 0
	basePos := 0

	for p = 0; p < PrefixArrLenfth; p++ {
		if n <= 0 {
			pfxLen++
			n = 1 << pfxLen
			basePos = n - 1
		}
		if IsBitSetInArr(t.prefixArr[:], p) == true {
			shftBits := TrieJmpLength - pfxLen
			pLevelPfxLen := level * TrieJmpLength
			cval := (p - basePos) << shftBits
			if p == 0 {
				pfxIdx = 0
			} else {
				pfxIdx = CountSetBitsInArr(t.prefixArr[:], p-1)
			}
			pfxStr = ""
			for i := 0; i < ts.maxLevels; i++ {
				var pfxVal = int(tv.prefix[i])
				var apStr string = "."
				if i == level {
					pfxVal = cval
				}
				if i == ts.maxLevels-1 {
					apStr = ""
				}
				pfxStr += fmt.Sprintf("%d%s", pfxVal, apStr)
			}
			td := tf.TrieData2String(t.prefixData[pfxIdx])
			tf.TrieNodeWalker(fmt.Sprintf("%20s/%d : %s", pfxStr, int(pfxLen)+pLevelPfxLen, td))
		}
		n--
	}
	for p = 0; p < PtrArrLength; p++ {
		if IsBitSetInArr(t.ptrArr[:], p) == true {
			cval := p
			ptrIdx := CountSetBitsInArr(t.ptrArr[:], p-1)

			if t.ptrData[ptrIdx] != nil {
				nextRoot := t.ptrData[ptrIdx]
				tv.prefix[level] = byte(cval)
				nextRoot.walkTrieInt(tv, level+1, ts, tf)
			}
		}
	}
	return 0
}

// AddTrie - Add a trie entry
// cidr is the route in cidr format and data is any user-defined data
// returns 0 on success or non-zero error code on error
func (t *TrieRoot) AddTrie(cidr string, data TrieData) int {
	var tv trieVar
	var ts = trieState{data, 0, 0, false, trieVar{}, false, 4, 0}

	pfxLen := cidr2TrieVar(cidr, &tv)

	if pfxLen < 0 {
		return TrieErrPrefix
	}

	ret := t.addTrieInt(&tv, 0, pfxLen, &ts)
	if ret != 0 || ts.errCode != 0 {
		return ret
	}

	return 0
}

// DelTrie - Delete a trie entry
// cidr is the route in cidr format
// returns 0 on success or non-zero error code on error
func (t *TrieRoot) DelTrie(cidr string) int {
	var tv trieVar
	var ts = trieState{0, 0, 0, false, trieVar{}, false, 4, 0}

	pfxLen := cidr2TrieVar(cidr, &tv)

	if pfxLen < 0 {
		return TrieErrPrefix
	}

	ret := t.deleteTrieInt(&tv, 0, pfxLen, &ts)
	if ret != 0 || ts.errCode != 0 {
		return TrieErrNoEnt
	}

	return 0
}

// FindTrie - Lookup matching route as per longest prefix match
// IP is the IP address in string format
// returns the following :
// 1. 0 on success or non-zero error code on error
// 2. matching route in *net.IPNet form
// 3. user-defined data associated with the trie entry
func (t *TrieRoot) FindTrie(IP string) (int, *net.IPNet, TrieData) {
	var tv trieVar
	var ts = trieState{0, 0, 0, false, trieVar{}, false, 4, 0}
	var cidr string

	if !t.v6 {
		cidr = IP + "/32"
	} else {
		cidr = IP + "/128"
	}
	pfxLen := cidr2TrieVar(cidr, &tv)

	if pfxLen < 0 {
		return TrieErrPrefix, nil, 0
	}

	t.findTrieInt(&tv, 0, &ts)

	if ts.matchFound {
		if !t.v6 {
			var res net.IP
			for i := 0; i < 4; i++ {
				res = append(res, ts.lastMatchTv.prefix[i])
			}
			mask := net.CIDRMask(ts.lastMatchPfxLen, 32)
			ipnet := net.IPNet{IP: res, Mask: mask}
			return 0, &ipnet, ts.trieData
		} else {
			var res net.IP = ts.lastMatchTv.prefix[:]
			mask := net.CIDRMask(ts.lastMatchPfxLen, 128)
			ipnet := net.IPNet{IP: res, Mask: mask}
			return 0, &ipnet, ts.trieData
		}
	}
	return TrieErrNoEnt, nil, 0
}

// Trie2String - stringify the trie table
func (t *TrieRoot) Trie2String(tf TrieIterIntf) {
	var ts = trieState{0, 0, 0, false, trieVar{}, false, 4, 0}
	t.walkTrieInt(&trieVar{}, 0, &ts, tf)
}
