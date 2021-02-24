package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type BackupData struct {
	Hosts         []string                `json:"hosts"`
	Notifications []StateChangeParams     `json:"notifications"`
	Checks        map[string][]ChecksData `json:"checks,omitempty"`
}

//Go max time.Time
//var maxTime time.Time = time.Unix(1<<63-62135596801, 999999999)
//PostgreSQL max timestamp value
var minTime time.Time = time.Unix(0, 0)
var maxTime time.Time = time.Unix(866459203200, 0)

const JsonBackupHandlerEndpoint string = "/api/backup"

func JsonBackupHandler(w http.ResponseWriter, r *http.Request) {
	if Config.Listen.WebAuth.Enable {
		username, password, authOK := r.BasicAuth()
		if authOK == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Not authorized"))
			return
		}

		if username != Config.Listen.WebAuth.User || password != Config.Listen.WebAuth.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Not authorized"))
			return
		}
	}

	var err error
	var buData BackupData

	if r.Method == http.MethodGet {
		buData.Hosts, err = MonData.GetHostsList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buData.Notifications, err = MonData.GetHostStateChangeParamsList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var jsonData []byte
		jsonData, err = json.Marshal(buData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var body []byte
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &buData)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		for _, h := range buData.Hosts {
			err = AddHost(h)
			if err != nil && err != ErrHostInDB {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		for _, n := range buData.Notifications {
			err = MonData.AddHostStateChangeParams(n.Host, n.ChangeThreshold, n.Action)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}

const JsonBackupFullHandlerEndpoint string = "/api/backup_full"

func JsonBackupFullHandler(w http.ResponseWriter, r *http.Request) {
	if Config.Listen.WebAuth.Enable {
		username, password, authOK := r.BasicAuth()
		if authOK == false {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Not authorized"))
			return
		}

		if username != Config.Listen.WebAuth.User || password != Config.Listen.WebAuth.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 - Not authorized"))
			return
		}
	}

	var err error
	var buData BackupData

	if r.Method == http.MethodGet {
		buData.Hosts, err = MonData.GetHostsList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buData.Notifications, err = MonData.GetHostStateChangeParamsList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		buData.Checks = make(map[string][]ChecksData)
		for _, h := range buData.Hosts {
			buData.Checks[h], err = MonData.GetChecksData(ChecksRequest{Host: h, Start: minTime, End: maxTime})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		var jsonData []byte
		jsonData, err = json.Marshal(buData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var body []byte
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &buData)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		for _, h := range buData.Hosts {
			err = AddHost(h)
			if err != nil && err != ErrHostInDB {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		for _, n := range buData.Notifications {
			err = MonData.AddHostStateChangeParams(n.Host, n.ChangeThreshold, n.Action)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if buData.Checks != nil {
			for k, v := range buData.Checks {
				for _, c := range v {
					err = MonData.SaveCheck(k, c.Timestamp.UTC(), c.Rtt, c.Up)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}
		}
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}
