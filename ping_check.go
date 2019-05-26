package main

import (
	"bytes"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"net/url"
	"os"
	"syscall"
	"time"
)

const protocolICMP = 1
const protocolIPv6ICMP = 58

func ipv4Payload(b []byte) []byte {
	if len(b) < ipv4.HeaderLen {
		return b
	}
	hdrlen := int(b[0]&0x0f) << 2
	return b[hdrlen:]
}

//rtt is in nanoseconds
//TODO: check checksumm
func Ping(host string, isV4 bool) (rtt int64, up bool, err error) {
	//Prevent panic
	defer func() {
		if r := recover(); r != nil {
			rtt = -1
			up = false
			err = r.(error)
			return
		}
	}()

	var c net.Conn
	if isV4 {
		c, err = net.Dial("ip4:icmp", host)
	} else {
		c, err = net.Dial("ip6:ipv6-icmp", host)
	}
	if err != nil {
		return -1, false, nil
	}

	err = c.SetDeadline(time.Now().Add(time.Duration(Config.Checks.Timeout) * time.Second))
	if err != nil {
		return -1, false, err
	}
	defer c.Close()

	xid := os.Getpid() & 0xffff
	xid = 0
	xseq := 0
	start := time.Now()
	sec := int64(start.Unix())
	usec := int64((start.UnixNano() / 1000 % 1000))
	xdata := int64ToBytes(sec)
	xdata = append(xdata, int64ToBytes(usec)...)
	//xdata = []byte("ping")

	var rq []byte
	if isV4 {
		rq, err = (&icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID: xid, Seq: xseq,
				Data: xdata,
			},
		}).Marshal(nil)
	} else {
		rq, err = (&icmp.Message{
			Type: ipv6.ICMPTypeEchoRequest,
			Code: 0,
			Body: &icmp.Echo{
				ID: xid, Seq: xseq,
				Data: xdata,
			},
		}).Marshal(nil)
	}

	if err != nil {
		return -1, false, err
	}

	if _, err := c.Write(rq); err != nil {
		return -1, false, err
	}

	rsp := make([]byte, 65535)
	n, err := c.Read(rsp)
	rtt = int64(time.Since(start))
	if err != nil {
		switch err := err.(type) {
		case *net.OpError:
			if err.Timeout() {
				return rtt, false, nil
			}
			if sysErr, ok := err.Err.(*os.SyscallError); ok {
				if errno, ok := sysErr.Err.(syscall.Errno); ok {
					if errno == syscall.ECONNABORTED || errno == syscall.ECONNRESET || errno == syscall.ECONNREFUSED || errno == syscall.ENETUNREACH || errno == syscall.EHOSTUNREACH || errno == syscall.EHOSTDOWN ||
						errno == syscall.Errno(10053) || errno == syscall.Errno(10054) || errno == syscall.Errno(10061) || errno == syscall.Errno(10060) {
						return rtt, false, nil
					}
				}
			}
		case *url.Error:
			if err, ok := err.Err.(net.Error); ok && err.Timeout() {
				return rtt, false, nil
			}
			if opErr, ok := err.Err.(*net.OpError); ok {
				if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
					if errno, ok := sysErr.Err.(syscall.Errno); ok {
						if errno == syscall.ECONNABORTED || errno == syscall.ECONNRESET || errno == syscall.ECONNREFUSED || errno == syscall.ENETUNREACH || errno == syscall.EHOSTUNREACH || errno == syscall.EHOSTDOWN ||
							errno == syscall.Errno(10053) || errno == syscall.Errno(10054) || errno == syscall.Errno(10061) || errno == syscall.Errno(10060) {
							return rtt, false, nil
						}
					}
				}
			}
		case net.Error:
			if err.Timeout() {
				return rtt, false, nil
			}
		}
		return rtt, false, err
	} else {
		rsp = rsp[:n]
	}

	if isV4 {
		rsp = ipv4Payload(rsp)
	}

	var m *icmp.Message
	if isV4 {
		if m, err = icmp.ParseMessage(protocolICMP, rsp); err != nil {
			return rtt, false, err
		}
	} else {
		if m, err = icmp.ParseMessage(protocolIPv6ICMP, rsp); err != nil {
			return rtt, false, err
		}
	}

	//invalid reply type check
	if isV4 {
		if m.Type != ipv4.ICMPTypeEchoReply {
			return rtt, false, nil
		}
	} else {
		if m.Type != ipv6.ICMPTypeEchoReply {
			return rtt, false, nil
		}
	}

	switch p := m.Body.(type) {
	case *icmp.Echo:
		if p.ID != xid {
			//invalid reply indentifier
			return rtt, false, nil
		}
		if p.Seq != xseq {
			//invalid reply sequence
			return rtt, false, nil
		}
		if !bytes.Equal(p.Data, xdata) {
			//invalid reply payload
			return rtt, false, nil
		}
		return rtt, true, nil
	default:
		//invalid reply
		return rtt, false, nil
	}
}
