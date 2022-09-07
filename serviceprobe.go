// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"net"
	"net/http"
	"time"
)

// HTTPProber - Do a http probe for given url
// returns true/false depending on whether probing was successful
func HTTPProber(urls string) bool {
	timeout := time.Duration(2 * time.Second)
	client := http.Client{Timeout: timeout}
	r, e := client.Head(urls)

	return e == nil && r.StatusCode == 200
}

// L4ServiceProber - Do a probe for L4 service end-points
// sType is "tcp" or "udp"
// sName is end-point IP address in string format
// returns true/false depending on whether probing was successful
func L4ServiceProber(sType string, sName string) bool {
	sOk := false
	timeout := 1 * time.Second

	if sType != "tcp" && sType != "udp" {
		// Unsupported
		return true
	}

	c, err := net.DialTimeout(sType, sName, timeout)
	if err != nil {
		sOk = false
	} else {
		sOk = true
	}
	if c != nil {
		defer c.Close()
	}

	return sOk
}
