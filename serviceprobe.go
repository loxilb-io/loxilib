// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"bytes"
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
// req is the request to be made to server (empty for none)
// resp is the response expected from server (empty for none)
// returns true/false depending on whether probing was successful
func L4ServiceProber(sType string, sName string, sHint, req, resp string) bool {
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

		cn, err := sctp.DialSCTP("sctp", laddr, addr, false)
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

	if req != "" && resp != "" {
		c.SetDeadline(time.Now().Add(2 * time.Second))
		_, err = c.Write([]byte(req))
		if err != nil {
			return false
		}
		aRb := []byte(resp)
		rRb := []byte(resp)
		_, err = c.Read(aRb)
		if err != nil {
			return false
		}
		if !bytes.Equal(aRb, rRb) {
			return false
		}
	}

	return sOk
}
