// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"math/bits"
)

func countAllSetBitsInArr(arr []uint8) int {
	var bCount int = 0
	sz := len(arr)

	for i := 0; i < sz; i++ {
		bCount += bits.OnesCount8(arr[i])
	}

	return bCount
}

func countSetBitsInArr(arr []uint8, bPos int) int {
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

func isBitSetInArr(arr []uint8, bPos int) bool {

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

func setBitInArr(arr []uint8, bPos int) {

	if int(bPos) >= 8*len(arr) {
		return
	}

	arrIdx := bPos / 8
	bPosIdx := 7 - (bPos % 8)
	arr[arrIdx] |= 0x1 << bPosIdx

	return
}

func unSetBitInArr(arr []uint8, bPos int) {

	if int(bPos) >= 8*len(arr) {
		return
	}

	arrIdx := bPos / 8
	bPosIdx := 7 - (bPos % 8)
	arr[arrIdx] &= ^(0x1 << bPosIdx)

	return
}
