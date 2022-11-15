// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"math/bits"
)

// CountAllSetBitsInArr - count set bits in an array of uint8
func CountAllSetBitsInArr(arr []uint8) int {
	var bCount int = 0
	sz := len(arr)

	for i := 0; i < sz; i++ {
		bCount += bits.OnesCount8(arr[i])
	}

	return bCount
}

// CountSetBitsInArr - count set bits in an array of uint8 from bPos
func CountSetBitsInArr(arr []uint8, bPos int) int {
	bCount := 0
	if int(bPos) >= 8*len(arr) {
		return -1
	}

	arrIdx := bPos / 8
	bPosIdx := 7 - (bPos % 8)

	for i := 0; i <= int(arrIdx); i++ {
		var val uint8
		if i == int(arrIdx) {
			val = arr[i] >> bPosIdx & 0xff

		} else {
			val = arr[i]
		}
		bCount += bits.OnesCount8(val)
	}

	return bCount
}

// IsBitSetInArr - check given bPos bit is set in the array
func IsBitSetInArr(arr []uint8, bPos int) bool {

	if int(bPos) >= 8*len(arr) {
		return false
	}

	arrIdx := bPos / 8
	bPosIdx := 7 - (bPos % 8)

	if (arr[arrIdx]>>bPosIdx)&0x1 == 0x1 {
		return true
	}

	return false
}

// SetBitInArr - set bPos bit in the array
func SetBitInArr(arr []uint8, bPos int) {

	if int(bPos) >= 8*len(arr) {
		return
	}

	arrIdx := bPos / 8
	bPosIdx := 7 - (bPos % 8)
	arr[arrIdx] |= 0x1 << bPosIdx

	return
}

// UnSetBitInArr - unset bPos bit in the array
func UnSetBitInArr(arr []uint8, bPos int) {

	if int(bPos) >= 8*len(arr) {
		return
	}

	arrIdx := bPos / 8
	bPosIdx := 7 - (bPos % 8)
	arr[arrIdx] &= ^(0x1 << bPosIdx)

	return
}
