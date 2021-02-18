package main

import (
	"database/sql"
	_ "modernc.org/ql/driver"
	"os"
	"time"
)

type MonDBQL struct {
	db *sql.DB
}

func (d *MonDBQL) Open(cfg Configuration) error {
	var err error

	var performInit bool = false
	if Config.DB.Database == "" || Config.DB.Type == "" {
		performInit = true
		d.db, err = sql.Open("ql-mem", "gosrvmon.db")
	} else {
		if _, err = os.Stat(Config.DB.Database); os.IsNotExist(err) {
			performInit = true
		}
		d.db, err = sql.Open("ql2", Config.DB.Database)
	}

	if err != nil {
		return err
	}
	if performInit {
		err = d.Init()
		if err != nil {
			return err
		}
	}

	if err = d.db.Ping(); err != nil {
		return err
	}

	return nil
}

func (d *MonDBQL) Close() error {
	err := d.db.Close()
	return err
}

func (d *MonDBQL) Init() error {
	tx, err := d.db.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
CREATE TABLE hosts
(
  host string NOT NULL
);

CREATE TABLE checks
(
  host string NOT NULL,
  check_time time NOT NULL,
  rtt int64 NOT NULL,
  up bool NOT NULL
);

CREATE TABLE state_change_params
(
  host string NOT NULL,
  change_threshold int64 NOT NULL,
  action string NOT NULL
);
`)

	if err != nil {
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *MonDBQL) GetHostsList() (hosts []string, err error) {
	return GetHostsListCommon(d.db)
}

func (d *MonDBQL) AddHost(newHost string) error {
	return AddHostCommon(d.db, newHost)
}

func (d *MonDBQL) DeleteHost(newHost string) error {
	return DeleteHostCommon(d.db, newHost)
}

func (d *MonDBQL) CheckHostExists(newHost string) error {
	return CheckHostExistsCommon(d.db, newHost)
}

func (d *MonDBQL) SaveCheck(host string, checkTime time.Time, rtt int64, up bool) error {
	return SaveCheckCommon(d.db, host, checkTime, rtt, up)
}

func (d *MonDBQL) GetChecksData(chkReq ChecksRequest) (cData []ChecksData, err error) {
	return GetChecksDataCommon(d.db, chkReq)
}

func (d *MonDBQL) GetLastCheckData(host string) (cData ChecksData, err error) {
	return GetLastCheckDataCommon(d.db, host)
}

func (d *MonDBQL) AddHostStateChangeParams(newHost string, newThreshold int64, newAction string) error {
	return AddHostStateChangeParamsCommon(d.db, newHost, newThreshold, newAction)
}

func (d *MonDBQL) GetHostStateChangeParams(host string) (p StateChangeParams, err error) {
	return GetHostStateChangeParamsCommon(d.db, host)
}

func (d *MonDBQL) GetHostStateChangeParamsList() (p []StateChangeParams, err error) {
	return GetHostStateChangeParamsListCommon(d.db)
}

func (d *MonDBQL) DeleteHostStateChangeParams(newHost string) error {
	return DeleteHostStateChangeParamsCommon(d.db, newHost)
}
