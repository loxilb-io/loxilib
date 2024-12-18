// SPDX-License-Identifier: Apache 2.0
// Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"bytes"
	"context"
	"encoding/binary"
	"github.com/loxilb-io/sctp"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SvcWait - Channel to wait for service reply
type SvcWait struct {
	wait chan bool
}

// SvcKey - Service Key
type SvcKey struct {
	Dst  string
	Port int
}

var (
	icmpRunner chan bool
	svcLock    sync.RWMutex
	svcs       map[SvcKey]*SvcWait
)

func waitForBoolChannelOrTimeout(ch <-chan bool, timeout time.Duration) (bool, bool) {
	select {
	case val := <-ch:
		return val, true
	case <-time.After(timeout):
		return false, false
	}
}

func listenForICMP6UNreachable() {
	conn, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
	if err != nil {
		// This can be due to ipv6 not enabled
		// In such a case probes for udp6 will not be available
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1500)
	//conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	icmpRunner <- true
	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Ignore timeout errors
			}
			continue
		}

		// Parse the ICMPv6 message
		message, err := icmp.ParseMessage(ipv6.ICMPTypeDestinationUnreachable.Protocol(), buffer[:n])
		if err != nil {
			continue
		}

		// Check for Destination Unreachable messages
		if message.Type == ipv6.ICMPTypeDestinationUnreachable {
			if body, ok := message.Body.(*icmp.DstUnreach); ok {
				pktData := body.Data
				if len(pktData) >= 48 {
					destIP := net.IP(pktData[24:40]).String()
					dport := int(pktData[42])<<8 | int(pktData[43])
					svcLock.Lock()
					key := SvcKey{Dst: destIP, Port: dport}
					if svcWait := svcs[key]; svcWait != nil {
						svcWait.wait <- true
					}
					svcLock.Unlock()
				}
			}
		}
	}
}

func listenForICMPUNreachable() {

	// Open a raw socket to listen for ICMP messages
	rc, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		os.Exit(1)
	}
	defer rc.Close()
	pktData := make([]byte, 1500)
	//rc.SetDeadline(time.Now().Add(5 * time.Second))
	icmpRunner <- true
	for {
		plen, _, err := rc.ReadFrom(pktData)
		if err != nil {
			continue
		}

		icmpNr, err := icmp.ParseMessage(1, pktData)
		if err != nil {
			continue
		}
		if icmpNr.Code == 3 && plen >= 8+20+8 {
			iph, err := ipv4.ParseHeader(pktData[8:])
			if err != nil {
				continue
			}

			if iph.Protocol == 17 {
				dport := int(binary.BigEndian.Uint16(pktData[30:32]))
				svcLock.Lock()
				key := SvcKey{Dst: iph.Dst.String(), Port: dport}
				if svcWait := svcs[key]; svcWait != nil {
					svcWait.wait <- true
				}
				svcLock.Unlock()
			}
		}
	}
}

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

	svcLock.Lock()
	if svcs == nil {
		icmpRunner = make(chan bool)
		svcs = map[SvcKey]*SvcWait{}
		go listenForICMPUNreachable()
		go listenForICMP6UNreachable()
		<-icmpRunner
		<-icmpRunner
	}
	svcLock.Unlock()

	sOk := false
	timeout := 1 * time.Second

	if sType != "tcp" && sType != "udp" && sType != "sctp" {
		// Unsupported
		return true
	}

	svcPair := strings.Split(sName, ":")
	if len(svcPair) < 2 {
		return false
	}

	portString := svcPair[len(svcPair)-1]
	svcPort, err := strconv.Atoi(portString)
	if err != nil || len(sName)-len(portString)-1 <= 0 {
		return false
	}

	netAddr := sName[:len(sName)-len(portString)-1]
	if netAddr[0:1] == "[" {
		netAddr = strings.Trim(netAddr, "[")
		netAddr = strings.Trim(netAddr, "]")
	}

	if sType == "sctp" {

		network := "ip4"
		if IsNetIPv6(netAddr) {
			network = "ip6"
		}

		epIp, err := net.ResolveIPAddr(network, netAddr)
		if err != nil {
			return false
		}

		ips := []net.IPAddr{*epIp}

		addr := &sctp.SCTPAddr{
			IPAddrs: ips,
			Port:    svcPort,
		}

		var laddr *sctp.SCTPAddr
		sIp, err := net.ResolveIPAddr(network, sHint)
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

	var c net.Conn
	if sHint == "" || sType == "udp" {
		c, err = net.DialTimeout(sType, sName, timeout)
	} else {
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP:   net.ParseIP(sHint),
				Port: 0,
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		c, err = dialer.DialContext(ctx, sType, sName)
	}
	if err != nil {
		sOk = false
	} else {
		sOk = true
	}
	if c != nil {
		defer c.Close()
	} else {
		return sOk
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
	} else if sType == "udp" {

		svcLock.Lock()
		key := SvcKey{Dst: netAddr, Port: svcPort}
		svcWait := svcs[key]
		if svcWait == nil {
			svcWait = &SvcWait{wait: make(chan bool)}
			svcs[key] = svcWait
		}
		svcLock.Unlock()

		c.SetDeadline(time.Now().Add(1 * time.Second))
		sOk = true
		_, err = c.Write([]byte("probe"))
		if err != nil {
			return false
		}
		pktData := make([]byte, 1500)
		_, err = c.Read(pktData)
		if err == nil {
			return sOk
		}

		period := 1 * time.Second
		if IsNetIPv6(netAddr) {
			period = 3 * time.Second
		}

		_, unRch := waitForBoolChannelOrTimeout(svcWait.wait, period)
		if unRch {
			sOk = false
		}

		svcLock.Lock()
		delete(svcs, key)
		svcLock.Unlock()
	}

	return sOk
}
