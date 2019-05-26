package main

import (
	"net"
	"net/url"
	"os"
	"syscall"
	"time"
)

func TcpCheck(host string) (rtt int64, up bool, err error) {
	network := "tcp"
	if false {
		network = "tcp4"
	} else if false {
		network = "tcp6"
	}

	start := time.Now()
	conn, err := net.DialTimeout(network, host, time.Duration(Config.Checks.Timeout)*time.Second)
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
	}
	err = conn.Close()

	return rtt, true, err
}
