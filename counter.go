// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"errors"
)

// Counter - context container
type Counter struct {
	begin    int
	end      int
	start    int
	len      int
	cap      int
	counters []int
}

// NewCounter - Allocate a set of counters
func NewCounter(begin int, length int) *Counter {
	counter := new(Counter)
	counter.counters = make([]int, length)
	counter.begin = begin
	counter.start = 0
	counter.end = length - 1
	counter.len = length
	counter.cap = length
	for i := 0; i < length; i++ {
		counter.counters[i] = i + 1
	}
	counter.counters[length-1] = -1
	return counter
}

// GetCounter - Get next available counter
func (C *Counter) GetCounter() (int, error) {
	if C.cap <= 0 || C.start == -1 {
		return -1, errors.New("Overflow")
	}

	C.cap--
	var rid = C.start
	if C.start == C.end {
		C.start = -1
	} else {
		C.start = C.counters[rid]
		C.counters[rid] = -1
	}
	return rid + C.begin, nil
}

// PutCounter - Return a counter to the available list
func (C *Counter) PutCounter(id int) error {
	if id < C.begin || id >= C.begin+C.len {
		return errors.New("Range")
	}
	rid := id - C.begin
	var tmp = C.end
	C.end = rid
	C.counters[tmp] = rid
	C.cap++
	if C.start == -1 {
		C.start = C.end
	}
	return nil
}
