package main

import (
	"bytes"
	"errors"
	"github.com/Alexander-r/bbolt"
	"time"
)

type MonDBBolt struct {
	db *bbolt.DB
}

func (d *MonDBBolt) Open(cfg Configuration) error {
	var err error
	d.db, err = bbolt.Open(cfg.DB.Database, 0666, &bbolt.Options{Timeout: 15 * time.Second})
	if err != nil {
		return err
	}
	err = d.Init()
	return err
}

func (d *MonDBBolt) Close() error {
	err := d.db.Close()
	return err
}

func (d *MonDBBolt) Init() error {
	err := d.db.Batch(func(tx *bbolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte("config:hosts"))
		if e != nil {
			return e
		}
		return nil
	})
	return err
}

func (d *MonDBBolt) GetHostsList() (hosts []string, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			kc := make([]byte, len(k))
			copy(kc, k)
			hosts = append(hosts, string(kc))
		}
		return nil
	})
	return hosts, err
}

func (d *MonDBBolt) AddHost(newHost string) error {
	err := d.db.Batch(func(tx *bbolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte(newHost))
		if e != nil {
			return e
		}
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return errors.New("DB not initialised")
		}
		var buf []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		e = b.Put([]byte(newHost), buf)
		return e
	})
	return err
}

func (d *MonDBBolt) DeleteHost(newHost string) error {
	err := d.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return errors.New("DB not initialised")
		}
		e := b.Delete([]byte(newHost))
		if e != nil {
			return e
		}
		e = tx.DeleteBucket([]byte(newHost))
		return nil
	})
	return err
}

func (d *MonDBBolt) CheckHostExists(newHost string) error {
	var hostExists bool = false
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(newHost))
		if v != nil {
			hostExists = true
		}
		return nil
	})
	if hostExists == false {
		return ErrNoHostInDB
	}
	return err
}

func (d *MonDBBolt) SaveCheck(host string, checkTime time.Time, rtt int64, up bool) error {
	err := d.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(host))
		if b == nil {
			return ErrNoHostInDB
		}
		var buf []byte = I64ToB(rtt)
		if up {
			buf = append(buf, 1)
		} else {
			buf = append(buf, 0)
		}
		t := I64ToB(checkTime.Unix())
		e := b.Put(t, buf)
		return e
	})
	return err
}

func (d *MonDBBolt) GetChecksData(chkReq ChecksRequest) (cData []ChecksData, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(chkReq.Host))
		if b == nil {
			return nil
		}
		tStart := I64ToB(chkReq.Start.Unix())
		tEnd := I64ToB(chkReq.End.Unix())
		c := b.Cursor()
		for k, v := c.Seek(tStart); k != nil && bytes.Compare(k, tEnd) <= 0; k, v = c.Next() {
			t := BToI64(k)
			if t == 0 {
				continue
			}
			if len(v) != 9 {
				continue
			}
			var up bool = false
			if v[8] != 0 {
				up = true
			}
			var cd ChecksData = ChecksData{Timestamp: time.Unix(t, 0).UTC(), Rtt: BToI64(v[:8]), Up: up}
			cData = append(cData, cd)
		}
		return nil
	})
	return cData, err
}

func (d *MonDBBolt) GetLastCheckData(host string) (cData ChecksData, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(host))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		k, v := c.Last()
		t := BToI64(k)
		if t == 0 {
			return nil
		}
		if len(v) != 9 {
			return nil
		}
		var up bool = false
		if v[8] != 0 {
			up = true
		}
		cData.Timestamp = time.Unix(t, 0).UTC()
		cData.Rtt = BToI64(v[:8])
		cData.Up = up
		return nil
	})
	return cData, err
}

func (d *MonDBBolt) AddHostStateChangeParams(newHost string, newThreshold int64, newAction string) error {
	var hostExists bool = false
	err := d.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return errors.New("DB not initialised")
		}
		v := b.Get([]byte(newHost))
		if v == nil {
			return nil
		}
		hostExists = true
		var buf []byte = I64ToB(newThreshold)
		buf = append(buf, []byte(newAction)...)
		e := b.Put([]byte(newHost), buf)
		return e
	})
	if hostExists == false && err == nil {
		return ErrNoHostInDB
	}
	return err
}

func (d *MonDBBolt) GetHostStateChangeParams(host string) (p StateChangeParams, err error) {
	var hostExists bool = false
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(host))
		if v != nil {
			if len(v) < 9 {
				return nil
			}
			hostExists = true
			p.ChangeThreshold = BToI64(v[:8])
			p.Action = string(v[8:])
			p.Host = host
		}
		return nil
	})
	if hostExists == false {
		return p, ErrNoHostInDB
	}
	return p, err
}

func (d *MonDBBolt) GetHostStateChangeParamsList() (p []StateChangeParams, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if len(v) < 9 {
				continue
			}
			var s StateChangeParams
			kc := make([]byte, len(k))
			copy(kc, k)
			vc := make([]byte, len(v[8:]))
			copy(vc, v[8:])
			s.ChangeThreshold = BToI64(v[:8])
			s.Action = string(vc)
			s.Host = string(kc)
			p = append(p, s)
		}
		return nil
	})
	return p, err
}

func (d *MonDBBolt) DeleteHostStateChangeParams(newHost string) error {
	err := d.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config:hosts"))
		if b == nil {
			return errors.New("DB not initialised")
		}
		v := b.Get([]byte(newHost))
		if v == nil {
			return nil
		}
		var buf []byte = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		e := b.Put([]byte(newHost), buf)
		return e
	})
	return err
}
