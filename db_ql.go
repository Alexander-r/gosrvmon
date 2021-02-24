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
	if cfg.DB.Database == "" || cfg.DB.Type == "" {
		performInit = true
		d.db, err = sql.Open("ql-mem", "gosrvmon.db")
	} else {
		if _, err = os.Stat(cfg.DB.Database); os.IsNotExist(err) {
			performInit = true
		}
		d.db, err = sql.Open("ql2", cfg.DB.Database)
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
  host int64 NOT NULL,
  check_time time NOT NULL,
  rtt int64 NOT NULL,
  up bool NOT NULL
);

CREATE INDEX checks_idx ON checks (host);

CREATE TABLE notifications_params
(
  host int64 NOT NULL,
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
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("DELETE FROM notifications_params WHERE host IN (SELECT id() FROM hosts WHERE host = $1 LIMIT 1);")
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
		stmt, err = tx.Prepare("DELETE FROM checks WHERE host IN (SELECT id() FROM hosts WHERE host = $1 LIMIT 1);")
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
		stmt, err = tx.Prepare("DELETE FROM hosts WHERE host=$1;")
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

func (d *MonDBQL) CheckHostExists(newHost string) error {
	return CheckHostExistsCommon(d.db, newHost)
}

func (d *MonDBQL) SaveCheck(host string, checkTime time.Time, rtt int64, up bool) error {
	var err error

	var tx *sql.Tx
	tx, err = d.db.Begin()
	if err != nil {
		return err
	}

	var stmt *sql.Stmt
	stmt, err = tx.Prepare("INSERT INTO checks (host, check_time, rtt, up) SELECT id(), $2, $3, $4 FROM hosts WHERE host = $1 LIMIT 1;")
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

func (d *MonDBQL) GetChecksData(chkReq ChecksRequest) (cData []ChecksData, err error) {
	cData = make([]ChecksData, 0)
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT check_time, rtt, up FROM checks WHERE host IN (SELECT id() FROM hosts WHERE host = $1 LIMIT 1) AND check_time >= $2 AND check_time <= $3;")
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

func (d *MonDBQL) GetLastCheckData(host string) (cData ChecksData, err error) {
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT check_time, rtt, up FROM checks WHERE host IN (SELECT id() FROM hosts WHERE host = $1 LIMIT 1) ORDER BY check_time DESC LIMIT 1;")
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

func (d *MonDBQL) DeleteOldChecks(beforeTime time.Time) error {
	return DeleteOldChecksCommon(d.db, beforeTime)
}

func (d *MonDBQL) AddHostStateChangeParams(newHost string, newThreshold int64, newAction string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("DELETE FROM notifications_params WHERE host IN (SELECT id() FROM hosts WHERE host = $1 LIMIT 1);")
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
		stmt, err = tx.Prepare("INSERT INTO notifications_params (host, change_threshold, action) SELECT id(), $2, $3 FROM hosts WHERE host = $1 LIMIT 1;")
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

func (d *MonDBQL) GetHostStateChangeParams(host string) (p StateChangeParams, err error) {
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT change_threshold, action FROM notifications_params WHERE host IN (SELECT id() FROM hosts WHERE host = $1 LIMIT 1);")
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

func (d *MonDBQL) GetHostStateChangeParamsList() (p []StateChangeParams, err error) {
	p = make([]StateChangeParams, 0)
	var stmt *sql.Stmt
	stmt, err = d.db.Prepare("SELECT hosts.host, notifications_params.change_threshold, notifications_params.action FROM hosts, notifications_params WHERE id(hosts) = notifications_params.host;")
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

func (d *MonDBQL) DeleteHostStateChangeParams(newHost string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("DELETE FROM notifications_params WHERE host IN (SELECT id() FROM hosts WHERE host = $1 LIMIT 1);")
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
