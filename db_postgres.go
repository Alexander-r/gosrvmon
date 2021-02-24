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
  id SERIAL NOT NULL,
  host text NOT NULL,
  UNIQUE(id),
  CONSTRAINT hosts_host_pkey PRIMARY KEY (host)
);

CREATE TABLE public.checks
(
  host integer NOT NULL,
  check_time timestamp without time zone NOT NULL,
  rtt bigint NOT NULL,
  up boolean NOT NULL,
  CONSTRAINT checks_pkey PRIMARY KEY (host, check_time),
  CONSTRAINT checks_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX checks_check_time_idx
  ON public.checks
  USING brin
  (host, check_time);

CREATE TABLE public.notifications_params
(
  host integer NOT NULL,
  change_threshold bigint NOT NULL,
  action text NOT NULL,
  CONSTRAINT notifications_params_pkey PRIMARY KEY (host),
  CONSTRAINT notifications_params_host_fkey FOREIGN KEY (host)
      REFERENCES public.hosts (id) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX notifications_params_host_idx
  ON public.notifications_params
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
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("DELETE FROM notifications_params WHERE host IN (SELECT id FROM hosts WHERE host = $1 LIMIT 1);")
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		_, err = stmt.Exec(newHost)
		if err != nil {
			stmt.Close()
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		err = stmt.Close()
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("DELETE FROM checks WHERE host IN (SELECT id FROM hosts WHERE host = $1 LIMIT 1);")
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		_, err = stmt.Exec(newHost)
		if err != nil {
			stmt.Close()
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		err = stmt.Close()
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("DELETE FROM hosts WHERE host = $1;")
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		_, err = stmt.Exec(newHost)
		if err != nil {
			stmt.Close()
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		err = stmt.Close()
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *MonDBPQ) CheckHostExists(newHost string) error {
	return CheckHostExistsCommon(d.db, newHost)
}

func (d *MonDBPQ) SaveCheck(host string, checkTime time.Time, rtt int64, up bool) error {
	var err error

	var tx *sql.Tx
	tx, err = d.db.Begin()
	if err != nil {
		return err
	}

	var stmt *sql.Stmt
	stmt, err = tx.Prepare("INSERT INTO checks (host, check_time, rtt, up) SELECT id, $2, $3, $4 FROM hosts WHERE host = $1 LIMIT 1;")
	if err != nil {
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}

	_, err = stmt.Exec(host, checkTime, rtt, up)
	if err != nil {
		stmt.Close()
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}

	err = stmt.Close()
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

func (d *MonDBPQ) GetChecksData(chkReq ChecksRequest) (cData []ChecksData, err error) {
	cData = make([]ChecksData, 0)
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT check_time, rtt, up FROM checks WHERE host IN (SELECT id FROM hosts WHERE host = $1 LIMIT 1) AND check_time >= $2 AND check_time <= $3;")
	if err != nil {
		return cData, err
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(chkReq.Host, chkReq.Start, chkReq.End)
	if err != nil {
		return cData, err
	}
	defer rows.Close()

	for rows.Next() {
		var tmpDat ChecksData
		err := rows.Scan(&tmpDat.Timestamp, &tmpDat.Rtt, &tmpDat.Up)
		if err != nil {
			return cData, err
		}
		cData = append(cData, tmpDat)
	}
	err = rows.Err()
	if err != nil {
		return cData, err
	}
	return cData, nil
}

func (d *MonDBPQ) GetLastCheckData(host string) (cData ChecksData, err error) {
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT check_time, rtt, up FROM checks WHERE host IN (SELECT id FROM hosts WHERE host = $1 LIMIT 1) ORDER BY check_time DESC LIMIT 1;")
	if err != nil {
		return cData, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(host)
	err = row.Scan(&cData.Timestamp, &cData.Rtt, &cData.Up)
	if err != nil {
		return cData, err
	}
	return cData, nil
}

func (d *MonDBPQ) DeleteOldChecks(beforeTime time.Time) error {
	return DeleteOldChecksCommon(d.db, beforeTime)
}

func (d *MonDBPQ) AddHostStateChangeParams(newHost string, newThreshold int64, newAction string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("DELETE FROM notifications_params WHERE host IN (SELECT id FROM hosts WHERE host = $1 LIMIT 1);")
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		_, err = stmt.Exec(newHost)
		if err != nil {
			stmt.Close()
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		err = stmt.Close()
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("INSERT INTO notifications_params (host, change_threshold, action) SELECT id, $2, $3 FROM hosts WHERE host = $1 LIMIT 1;")
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		_, err = stmt.Exec(newHost, newThreshold, newAction)
		if err != nil {
			stmt.Close()
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}

		err = stmt.Close()
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				return e
			}
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *MonDBPQ) GetHostStateChangeParams(host string) (p StateChangeParams, err error) {
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT change_threshold, action FROM notifications_params WHERE host IN (SELECT id FROM hosts WHERE host = $1 LIMIT 1);")
	if err != nil {
		return p, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(host)
	err = row.Scan(&p.ChangeThreshold, &p.Action)
	p.Host = host
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoHostInDB
		}
		return p, err
	}
	return p, nil
}

func (d *MonDBPQ) GetHostStateChangeParamsList() (p []StateChangeParams, err error) {
	p = make([]StateChangeParams, 0)
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT hosts.host, notifications_params.change_threshold, notifications_params.action FROM hosts, notifications_params WHERE hosts.id = notifications_params.host;")
	if err != nil {
		return p, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		return p, err
	}
	defer rows.Close()
	for rows.Next() {
		var s StateChangeParams
		err = rows.Scan(&s.Host, &s.ChangeThreshold, &s.Action)
		if err != nil {
			return p, err
		}
		p = append(p, s)
	}
	err = rows.Err()
	if err != nil {
		return p, err
	}
	return p, nil
}

func (d *MonDBPQ) DeleteHostStateChangeParams(newHost string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("DELETE FROM notifications_params WHERE host IN (SELECT id FROM hosts WHERE host = $1 LIMIT 1);")
	if err != nil {
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}

	_, err = stmt.Exec(newHost)
	if err != nil {
		stmt.Close()
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}

	err = stmt.Close()
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
