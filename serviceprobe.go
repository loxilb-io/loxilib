// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"github.com/loxilb-io/sctp"
	"net"
	"net/http"
	"strconv"
	"strings"
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
// sType is "tcp" or "udp" or "sctp"
// sName is end-point IP address in string format
// sHint is source address hint if any
// returns true/false depending on whether probing was successful
func L4ServiceProber(sType string, sName string, sHint string) bool {
	sOk := false
	timeout := 1 * time.Second

	if sType != "tcp" && sType != "udp" && sType != "sctp" {
		// Unsupported
		return true
	}

	if sType == "sctp" {
		svcPair := strings.Split(sName, ":")
		if len(svcPair) != 2 {
			return false
		}

		svcPort, err := strconv.Atoi(svcPair[1])
		if err != nil {
			return false
		}

		epIp, err := net.ResolveIPAddr("ip", svcPair[0])
		if err != nil {
			return false
		}

		ips := []net.IPAddr{*epIp}

		addr := &sctp.SCTPAddr{
			IPAddrs: ips,
			Port:    svcPort,
		}

		var laddr *sctp.SCTPAddr
		sIp, err := net.ResolveIPAddr("ip", sHint)
		if err == nil {
			sips := []net.IPAddr{*sIp}
			laddr = &sctp.SCTPAddr{
				IPAddrs: sips,
				Port:    12346,
			}
		}

		cn, err := sctp.DialSCTP("sctp", laddr, addr, true)
		if err != nil {
			sOk = false
		} else {
			sOk = true
		}

		if cn != nil {
			cn.Close()
		}

		return sOk
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
