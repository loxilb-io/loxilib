// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"errors"
)

// Counter - context container
type Counter struct {
	begin    uint64
	end      uint64
	start    uint64
	len      uint64
	cap      uint64
	counters []uint64
}

// NewCounter - Allocate a set of counters
func NewCounter(begin uint64, length uint64) *Counter {
	counter := new(Counter)
	counter.counters = make([]uint64, length)
	counter.begin = begin
	counter.start = 0
	counter.end = length - 1
	counter.len = length
	counter.cap = length
	for i := uint64(0); i < length; i++ {
		counter.counters[i] = i + 1
	}
	counter.counters[length-1] = ^uint64(0)
	return counter
}

// GetCounter - Get next available counter
func (C *Counter) GetCounter() (uint64, error) {
	if C.cap <= 0 || C.start == ^uint64(0) {
		return ^uint64(0), errors.New("Overflow")
	}

	C.cap--
	var rid = C.start
	if C.start == C.end {
		C.start = ^uint64(0)
	} else {
		C.start = C.counters[rid]
		C.counters[rid] = ^uint64(0)
	}
	return rid + C.begin, nil
}

// PutCounter - Return a counter to the available list
func (C *Counter) PutCounter(id uint64) error {
	if id < C.begin || id >= C.begin+C.len {
		return errors.New("Range")
	}
	rid := id - C.begin
	var tmp = C.end
	C.end = rid
	C.counters[tmp] = rid
	C.cap++
	if C.start == ^uint64(0) {
		C.start = C.end
	}
	return nil
}

// ReserveCounter - Don't allocate this counter
func (C *Counter) ReserveCounter(id uint64) error {
	if id < C.begin || id >= C.begin+C.len {
		return errors.New("Range")
	}

	if C.cap <= 0 || C.start == ^uint64(0) || C.counters[id] == ^uint64(0) {
		return errors.New("Overflow")
	}

	tmp := C.start
	C.start = C.counters[id]
	C.end = tmp
	C.counters[id] = ^uint64(0)
	C.cap--

	return nil
}
