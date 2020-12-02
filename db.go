package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	_ "modernc.org/ql/driver"
	"os"
	"strings"
	"time"
)

var db *sql.DB

func dBConnect() error {
	var err error

	switch Config.DB.Type {
	case "pq":
		var connStr strings.Builder

		connStr.WriteString("host=")
		connStr.WriteString(Config.DB.Host)
		connStr.WriteString(" port=")
		connStr.WriteString(Config.DB.Port)
		connStr.WriteString(" dbname=")
		connStr.WriteString(Config.DB.Database)
		connStr.WriteString(" user=")
		connStr.WriteString(Config.DB.User)
		connStr.WriteString(" password=")
		connStr.WriteString(Config.DB.Password)
		connStr.WriteString(" sslmode=disable TimeZone=utc connect_timeout=120")

		db, err = sql.Open("postgres", connStr.String())
	case "ql":
		var performInit bool = false
		if _, err = os.Stat(Config.DB.Database); os.IsNotExist(err) {
			performInit = true
		}
		db, err = sql.Open("ql2", Config.DB.Database)
		if err != nil {
			return err
		}
		if performInit {
			err = dBInit()
		}
	default:
		db, err = sql.Open("ql-mem", "gosrvmon.db")
		if err != nil {
			return err
		}
		err = dBInit()
	}

	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	return nil
}

func dBClose() {
	err := db.Close()
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}

func dBInit() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	switch Config.DB.Type {
	case "pq":
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
	//case "ql": //NOTE: same as default
	default:
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
	}

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

func getHostsList() (hosts []string, err error) {
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

func addHost(newHost string) error {
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

func deleteHost(newHost string) error {
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

func checkHostExists(newHost string) error {
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

func saveCheck(host string, checkTime time.Time, rtt int64, up bool) error {
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

func getChecksData(chkReq ChecksRequest) (cData []ChecksData, err error) {
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

func getLastCheckData(host string) (cData ChecksData, err error) {
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

func addHostStateChangeParams(newHost string, newThreshold int64, newAction string) error {
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

func getHostStateChangeParams(host string) (p StateChangeParams, err error) {
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

func getHostStateChangeParamsList() (p []StateChangeParams, err error) {
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

func deleteHostStateChangeParams(newHost string) error {
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
