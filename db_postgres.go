package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

type MonDBPQ struct {
	db *sql.DB
}

func (d *MonDBPQ) Open(cfg Configuration) error {
	var err error

	var connStr strings.Builder

	connStr.WriteString("host=")
	connStr.WriteString(cfg.DB.Host)
	connStr.WriteString(" port=")
	connStr.WriteString(cfg.DB.Port)
	connStr.WriteString(" dbname=")
	connStr.WriteString(cfg.DB.Database)
	connStr.WriteString(" user=")
	connStr.WriteString(cfg.DB.User)
	connStr.WriteString(" password=")
	connStr.WriteString(cfg.DB.Password)
	connStr.WriteString(" sslmode=disable TimeZone=utc connect_timeout=120")

	d.db, err = sql.Open("postgres", connStr.String())

	if err != nil {
		return err
	}

	if err = d.db.Ping(); err != nil {
		return err
	}

	return nil
}

func (d *MonDBPQ) Close() error {
	err := d.db.Close()
	return err
}

func (d *MonDBPQ) Init() error {
	tx, err := d.db.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
CREATE TABLE public.hosts
(
  host text NOT NULL,
  CONSTRAINT hosts_host_pkey PRIMARY KEY (host)
);

CREATE TABLE public.checks
(
  host text NOT NULL,
  check_time timestamp without time zone NOT NULL,
  rtt bigint NOT NULL,
  up boolean NOT NULL,
  CONSTRAINT checks_pkey PRIMARY KEY (host, check_time),
  CONSTRAINT checks_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (host) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX checks_check_time_idx
  ON public.checks
  USING brin
  (host, check_time);

CREATE TABLE public.state_change_params
(
  host text NOT NULL,
  change_threshold bigint NOT NULL,
  action text NOT NULL,
  CONSTRAINT state_change_params_pkey PRIMARY KEY (host),
  CONSTRAINT state_change_params_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (host) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX state_change_params_host_idx
  ON public.state_change_params
  USING brin
  (host);
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

func (d *MonDBPQ) GetHostsList() (hosts []string, err error) {
	return GetHostsListCommon(d.db)
}

func (d *MonDBPQ) AddHost(newHost string) error {
	return AddHostCommon(d.db, newHost)
}

func (d *MonDBPQ) DeleteHost(newHost string) error {
	return DeleteHostCommon(d.db, newHost)
}

func (d *MonDBPQ) CheckHostExists(newHost string) error {
	return CheckHostExistsCommon(d.db, newHost)
}

func (d *MonDBPQ) SaveCheck(host string, checkTime time.Time, rtt int64, up bool) error {
	return SaveCheckCommon(d.db, host, checkTime, rtt, up)
}

func (d *MonDBPQ) GetChecksData(chkReq ChecksRequest) (cData []ChecksData, err error) {
	return GetChecksDataCommon(d.db, chkReq)
}

func (d *MonDBPQ) GetLastCheckData(host string) (cData ChecksData, err error) {
	return GetLastCheckDataCommon(d.db, host)
}

func (d *MonDBPQ) AddHostStateChangeParams(newHost string, newThreshold int64, newAction string) error {
	return AddHostStateChangeParamsCommon(d.db, newHost, newThreshold, newAction)
}

func (d *MonDBPQ) GetHostStateChangeParams(host string) (p StateChangeParams, err error) {
	return GetHostStateChangeParamsCommon(d.db, host)
}

func (d *MonDBPQ) GetHostStateChangeParamsList() (p []StateChangeParams, err error) {
	return GetHostStateChangeParamsListCommon(d.db)
}

func (d *MonDBPQ) DeleteHostStateChangeParams(newHost string) error {
	return DeleteHostStateChangeParamsCommon(d.db, newHost)
}
