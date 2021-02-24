package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var MonData MonDB

type CheckType int32

const (
	checkInvalid CheckType = -1
	checkIcmp    CheckType = 0
	checkHttp    CheckType = 1
	checkTcp     CheckType = 2
)

var (
	configPath = flag.String("config", "", "Config path")
	configStr  = flag.String("confstr", "", "Pass config as a string")
	checkHost  = flag.String("check", "", "Check single host")
	initDB     = flag.Bool("init", false, "Init DB")
)

var wg sync.WaitGroup

var doProcess bool = true

var hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

func isHostOrIP(host string) bool {
	if len(host) > 256 {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		if !hostnameRegex.MatchString(host) {
			return false
		}
	}
	return true
}

func getCheckType(host string) CheckType {
	if len(host) == 0 {
		return checkInvalid
	}
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		_, err := url.ParseRequestURI(host)
		if err != nil {
			return checkInvalid
		} else {
			return checkHttp
		}
	}
	var h []string
	if strings.HasPrefix(host, "[") {
		h = strings.Split(host[1:], "]:")
	} else {
		h = strings.Split(host, ":")
	}
	if len(h) == 2 {
		if isHostOrIP(h[0]) {
			if i, err := strconv.ParseInt(h[1], 10, 32); err == nil {
				if i > 0 && i <= 65536 {
					return checkTcp
				}
			}
		}
		return checkInvalid
	}
	if isHostOrIP(host) {
		return checkIcmp
	}
	return checkInvalid
}

func doCheck(host string, checkTime time.Time) {
	var rtt int64
	var up bool
	var err error

	checkType := getCheckType(host)

	switch checkType {
	case checkIcmp:
		rtt, up, err = PingCheck(host)
	case checkHttp:
		rtt, up, err = HttpCheck(host, Config.Checks.HTTPMethod)
	case checkTcp:
		rtt, up, err = TcpCheck(host)
	default:
		log.Println("[ERROR] Unknown checkType")
		return
	}

	if err != nil {
		up = false
	}
	go checkStateChange(host, rtt, checkTime, up)
	err = MonData.SaveCheck(host, checkTime, rtt, up)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
	wg.Done()
}

func checkTick(t time.Time) {
	if !doProcess {
		return
	}
	hosts, err := MonData.GetHostsList()
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return
	}
	for _, host := range hosts {
		wg.Add(1)
		go doCheck(host, t)
	}
}

func doSingleCheck(host string) (bool, time.Duration, error) {
	var rtt int64
	var up bool
	var err error

	checkType := getCheckType(host)

	switch checkType {
	case checkIcmp:
		rtt, up, err = PingCheck(host)
	case checkHttp:
		rtt, up, err = HttpCheck(host, Config.Checks.HTTPMethod)
	case checkTcp:
		rtt, up, err = TcpCheck(host)
	default:
		log.Println("[ERROR] Unknown checkType")
	}

	rttDuration := time.Duration(rtt)

	return up, rttDuration, err
}

var ctx, ctxCancel = context.WithCancel(context.Background())
var shutdownChan = make(chan struct{})

func Wait(duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-shutdownChan:
		return
	case <-time.After(duration):
		return
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	CheckByteOrder()

	var err error

	err = loadConfiguration()
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return
	}

	if len(*checkHost) > 0 {
		var up bool
		var rtt time.Duration
		up, rtt, err = doSingleCheck(*checkHost)
		if err == nil {
			log.Printf("Up: %v, rtt: %v", up, rtt)
		} else {
			log.Printf("[ERROR] %v", err)
		}
		return
	}

	switch Config.DB.Type {
	case "bolt":
		MonData = &MonDBBolt{}
	case "pq":
		MonData = &MonDBPQ{}
	case "ql":
		MonData = &MonDBQL{}
	default:
		MonData = &MonDBBolt{}
	}

	err = MonData.Open(Config)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return
	}
	defer MonData.Close()

	if *initDB {
		err = MonData.Init()
		if err != nil {
			log.Printf("[ERROR] %v", err)
			return
		}
		return
	}

	http.HandleFunc(IndexTemplateHandlerRootEndpoint, IndexTemplateHandler)
	http.HandleFunc(IndexTemplateHandlerHtmlEndpoint, IndexTemplateHandler)
	http.HandleFunc(JsonHostsHandlerEndpoint, JsonHostsHandler)
	http.HandleFunc(JsonChecksHandlerEndpoint, JsonChecksHandler)
	http.HandleFunc(JsonChecksLastHandlerEndpoint, JsonChecksLastHandler)
	http.HandleFunc(HostsTemplateHandlerEndpoint, HostsTemplateHandler)
	http.HandleFunc(ChecksTemplateHandlerEndpoint, ChecksTemplateHandler)
	http.HandleFunc(HostsViewTemplateHandlerEndpoint, HostsViewTemplateHandler)
	http.HandleFunc(ChecksChartEndpoint, checksChart)
	http.HandleFunc(StateChangeParamsHandlerEndpoint, StateChangeParamsTemplateHandler)
	http.HandleFunc(JsonStateChangeParamsHandlerEndpoint, JsonStateChangeParamsHandler)
	http.HandleFunc(JsonBackupHandlerEndpoint, JsonBackupHandler)
	http.HandleFunc(JsonBackupFullHandlerEndpoint, JsonBackupFullHandler)
	http.HandleFunc("/favicon.ico", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte{})
	})
	http.HandleFunc(flatpickrCssEndpoint, func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/css")
		res.Write([]byte(flatpickrCss))
	})
	http.HandleFunc(flatpickrJsEndpoint, func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		res.Write([]byte(flatpickrJs))
	})
	if Config.Checks.AllowSingleChecks {
		http.HandleFunc(JsonCheckHandlerEndpoint, JsonCheckHandler)
	}

	server := &http.Server{
		Addr:         Config.Listen.Address + ":" + Config.Listen.Port,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  time.Duration(Config.Listen.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(Config.Listen.WriteTimeout) * time.Second,
	}
	go func() {
		e := server.ListenAndServe()
		if e != nil {
			if e != http.ErrServerClosed {
				log.Printf("[ERROR] %v", e)
			}
			doProcess = false
		}
	}()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		doProcess = false
		close(shutdownChan)
		ctxCancel()
		wg.Wait()
		e := server.Close()
		if e != nil {
			log.Printf("[ERROR] %v", e)
		}
	}()

	dt := time.Duration(Config.Checks.Interval) * time.Second
	retentiont := time.Duration(-Config.Checks.Retention) * time.Second
	for doProcess {
		t := time.Now()
		if Config.Checks.Retention != 0 {
			go func() {
				e := MonData.DeleteOldChecks(t.Truncate(dt).Add(retentiont))
				if e != nil {
					log.Printf("[ERROR] %v", e)
				}
			}()
		}
		n := t.Truncate(dt).Add(dt)
		d := n.Sub(t)
		Wait(d)
		if Config.Checks.PerformChecks {
			go checkTick(n.UTC())
		}
	}
}
