package main

import (
	"errors"
	"log"
	"net"
	"strings"
	"sync"
)

type PingCheckResult struct {
	rtt int64
	up  bool
	err error
}

func IsIPv4(ad string) bool {
	return len(ad) > 0 && !strings.Contains(ad, ":")
}

func PingCheck(host string) (rtt int64, up bool, err error) {
	var addr []net.IP
	addr, err = net.LookupIP(host)
	if err != nil {
		return -1, false, nil
	}
	if len(addr) == 0 {
		return -1, false, errors.New("host has no A/AAAA records")
	}
	rtt = -1
	up = false
	err = nil
	for i := uint32(0); i < Config.Checks.PingRetryCount; i++ {
		c := make(chan PingCheckResult)
		go func() {
			var wg sync.WaitGroup
			for _, ad := range addr {
				if ad != nil {
					adstr := ad.String()
					if IsIPv4(adstr) {
						wg.Add(1)
						go func() {
							var pingRes PingCheckResult
							pingRes.rtt, pingRes.up, pingRes.err = Ping(adstr, true)
							c <- pingRes
							wg.Done()
						}()
					} else {
						wg.Add(1)
						go func() {
							var pingRes PingCheckResult
							pingRes.rtt, pingRes.up, pingRes.err = Ping(adstr, false)
							c <- pingRes
							wg.Done()
						}()
					}
				}
			}
			wg.Wait()
			close(c)
		}()
		for pingRes := range c {
			if pingRes.err != nil {
				log.Printf("[ERROR] %v", pingRes.err)
			} else {
				if pingRes.up {
					if !up {
						up = pingRes.up
						rtt = pingRes.rtt
					} else if pingRes.rtt < rtt {
						rtt = pingRes.rtt
					}
				}
			}
		}
	}
	return
}
