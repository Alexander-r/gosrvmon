package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

func GetCheckRequest(w http.ResponseWriter, r *http.Request) (chkHost string, err error) {
	switch r.Method {
	case http.MethodGet:
		chkHost = r.URL.Query().Get("host")

	case http.MethodPost:
		var body []byte
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		chkHost = string(body)

	default:
		err = errors.New("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(chkHost) <= 0 {
		err = errors.New("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if getCheckType(chkHost) == checkInvalid {
		err = errors.New("Host not acceptable")
		http.Error(w, "Host not acceptable", http.StatusBadRequest)
		return
	}

	return chkHost, nil
}

const JsonCheckHandlerEndpoint string = "/api/check"

func JsonCheckHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var chkHost string
	var rtt time.Duration
	chkHost, err = GetCheckRequest(w, r)
	if err != nil {
		return
	}

	var cData ChecksData
	cData.Timestamp = time.Now().UTC()

	cData.Up, rtt, err = doSingleCheck(chkHost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cData.Rtt = rtt.Nanoseconds()

	var jsonData []byte
	jsonData, err = json.Marshal(cData)
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
}
