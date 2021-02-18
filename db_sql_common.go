package main

import (
	"database/sql"
	"time"
)

func GetHostsListCommon(db *sql.DB) (hosts []string, err error) {
	hosts = make([]string, 0)
	var stmt *sql.Stmt
	stmt, err = db.Prepare("SELECT host FROM hosts;")
	if err != nil {
		return hosts, err
	}
	defer stmt.Close()
	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		return hosts, err
	}
	defer rows.Close()
	for rows.Next() {
		var tmpHost string
		err = rows.Scan(&tmpHost)
		if err != nil {
			return hosts, err
		}
		hosts = append(hosts, tmpHost)
	}
	err = rows.Err()
	if err != nil {
		return hosts, err
	}
	return hosts, nil
}

func AddHostCommon(db *sql.DB, newHost string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("INSERT INTO hosts (host) VALUES ($1);")
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

func DeleteHostCommon(db *sql.DB, newHost string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
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

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func CheckHostExistsCommon(db *sql.DB, newHost string) error {
	stmt, err := db.Prepare("SELECT host FROM hosts WHERE host=$1;")
	if err != nil {
		return err
	}
	defer stmt.Close()
	row := stmt.QueryRow(newHost)
	var tmpHost string
	err = row.Scan(&tmpHost)
	if err != nil {
		return err
	}
	return nil
}

func SaveCheckCommon(db *sql.DB, host string, checkTime time.Time, rtt int64, up bool) error {
	var err error

	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		return err
	}

	var stmt *sql.Stmt
	stmt, err = tx.Prepare("INSERT INTO checks (host, check_time, rtt, up) VALUES ($1, $2, $3, $4);")
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

func GetChecksDataCommon(db *sql.DB, chkReq ChecksRequest) (cData []ChecksData, err error) {
	cData = make([]ChecksData, 0)
	var stmt *sql.Stmt
	stmt, err = db.Prepare("SELECT check_time, rtt, up FROM checks WHERE host = $1 AND check_time >= $2 and check_time <= $3;")
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

func GetLastCheckDataCommon(db *sql.DB, host string) (cData ChecksData, err error) {
	var stmt *sql.Stmt
	stmt, err = db.Prepare("SELECT check_time, rtt, up FROM checks WHERE host = $1 ORDER BY check_time DESC LIMIT 1;")
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

func AddHostStateChangeParamsCommon(db *sql.DB, newHost string, newThreshold int64, newAction string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	{
		var stmt *sql.Stmt
		stmt, err = tx.Prepare("DELETE FROM state_change_params WHERE host=$1;")
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
		stmt, err = tx.Prepare("INSERT INTO state_change_params (host, change_threshold, action) VALUES ($1, $2, $3);")
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

func GetHostStateChangeParamsCommon(db *sql.DB, host string) (p StateChangeParams, err error) {
	var stmt *sql.Stmt
	stmt, err = db.Prepare("SELECT host, change_threshold, action FROM state_change_params WHERE host = $1;")
	if err != nil {
		return p, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(host)
	err = row.Scan(&p.Host, &p.ChangeThreshold, &p.Action)
	if err != nil {
		return p, err
	}
	return p, nil
}

func GetHostStateChangeParamsListCommon(db *sql.DB) (p []StateChangeParams, err error) {
	p = make([]StateChangeParams, 0)
	var stmt *sql.Stmt
	stmt, err = db.Prepare("SELECT host, change_threshold, action FROM state_change_params;")
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

func DeleteHostStateChangeParamsCommon(db *sql.DB, newHost string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("DELETE FROM state_change_params WHERE host=$1;")
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
