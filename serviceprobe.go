// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"net"
	"net/http"
	"time"
)

func HttpProber(urls string) bool {
	timeout := time.Duration(2 * time.Second)
	client := http.Client{Timeout: timeout}
	r, e := client.Head(urls)

	return e == nil && r.StatusCode == 200
}

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
