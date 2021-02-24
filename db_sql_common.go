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
		if err == sql.ErrNoRows {
			err = ErrNoHostInDB
		}
		return err
	}
	return nil
}

func DeleteOldChecksCommon(db *sql.DB, beforeTime time.Time) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("DELETE FROM checks WHERE check_time < $1;")
	if err != nil {
		e := tx.Rollback()
		if e != nil {
			return e
		}
		return err
	}

	_, err = stmt.Exec(beforeTime)
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
