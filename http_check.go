package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"
)

func HttpCheck(targetUrl string, metod string) (rtt int64, up bool, err error) {
	client := &http.Client{
		Timeout: time.Duration(Config.Checks.Timeout) * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
			//TLSClientConfig: &tls.Config{
			//	InsecureSkipVerify: true,
			//},
		},
	}

	var req *http.Request
	req, err = http.NewRequest(metod, targetUrl, nil)
	if err != nil {
		return rtt, false, err
	}
	req = req.WithContext(ctx)

	start := time.Now()
	var resp *http.Response
	resp, err = client.Do(req)

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

	if resp != nil {
		err = resp.Body.Close()
		if err != nil {
			return rtt, false, err
		}
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 399) {
		return rtt, false, fmt.Errorf("Response status: %v", resp.StatusCode)
	}

	return rtt, true, nil
}
