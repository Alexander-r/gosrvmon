package main

import (
	"log"
	"sync"
	"time"
)

type StateChangeParams struct {
	Host            string `json:"host"`
	ChangeThreshold int64  `json:"threshold"`
	Action          string `json:"action"`
}

type StateChangeData struct {
	LastTimeObserved time.Time `json:"observed"`
	State            bool      `json:"state"`
}

var CheckStates map[string]StateChangeData = make(map[string]StateChangeData)
var CheckStatesMux sync.RWMutex

func checkStateChange(host string, rtt int64, checkTime time.Time, up bool) {
	checkParams, err := MonData.GetHostStateChangeParams(host)
	if err != nil {
		if err != ErrNoHostInDB {
			log.Printf("[ERROR] %v", err)
		}
		return
	}
	CheckStatesMux.RLock()
	checkState, ok := CheckStates[host]
	CheckStatesMux.RUnlock()
	if !ok {
		CheckStatesMux.Lock()
		CheckStates[host] = StateChangeData{checkTime, up}
		CheckStatesMux.Unlock()
		return
	}
	if checkState.State == up {
		CheckStatesMux.Lock()
		CheckStates[host] = StateChangeData{checkTime, checkState.State}
		CheckStatesMux.Unlock()
		return
	}
	if checkTime.Sub(checkState.LastTimeObserved).Milliseconds()/1000 > checkParams.ChangeThreshold {
		CheckStatesMux.Lock()
		CheckStates[host] = StateChangeData{checkTime, up}
		CheckStatesMux.Unlock()
		err = EventHTTPNotify(host, rtt, checkTime, up, checkParams.Action)
		if err != nil {
			log.Printf("[ERROR] %v", err)
		}
	}
}
