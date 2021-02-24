package main

import (
	"errors"
	"time"
)

type MonDB interface {
	Open(cfg Configuration) error
	Close() error
	Init() error
	GetHostsList() (hosts []string, err error)
	AddHost(newHost string) error
	DeleteHost(newHost string) error
	CheckHostExists(newHost string) error
	SaveCheck(host string, checkTime time.Time, rtt int64, up bool) error
	GetChecksData(chkReq ChecksRequest) (cData []ChecksData, err error)
	GetLastCheckData(host string) (cData ChecksData, err error)
	DeleteOldChecks(beforeTime time.Time) error
	AddHostStateChangeParams(newHost string, newThreshold int64, newAction string) error
	GetHostStateChangeParams(host string) (p StateChangeParams, err error)
	GetHostStateChangeParamsList() (p []StateChangeParams, err error)
	DeleteHostStateChangeParams(newHost string) error
}

var ErrNoHostInDB = errors.New("no such host in DB")
