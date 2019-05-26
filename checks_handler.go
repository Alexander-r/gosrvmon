package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type ChecksData struct {
	Timestamp time.Time `json:"time"`
	Rtt       int64     `json:"rtt"`
	Up        bool      `json:"up"`
}

type ChecksRequest struct {
	Host  string    `json:"host"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func GetChecksRequest(w http.ResponseWriter, r *http.Request) (chkReq ChecksRequest, err error) {
	switch r.Method {
	case http.MethodGet:
		chkReq.Host = r.URL.Query().Get("host")
		startT := r.URL.Query().Get("start")
		endT := r.URL.Query().Get("end")

		if len(endT) <= 0 {
			chkReq.End = time.Now().UTC()
		} else {
			var endTime int64
			endTime, err = strconv.ParseInt(endT, 10, 64)

			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			chkReq.End = time.Unix(endTime, 0).UTC()
		}

		if len(startT) <= 0 {
			chkReq.Start = chkReq.End.Add(-24 * time.Hour).UTC()
		} else {
			var startTime int64
			startTime, err = strconv.ParseInt(startT, 10, 64)

			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			chkReq.Start = time.Unix(startTime, 0).UTC()
		}

	case http.MethodPost:
		var body []byte
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &chkReq)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

	default:
		err = errors.New("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(chkReq.Host) <= 0 {
		err = errors.New("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if chkReq.End.Before(chkReq.Start) {
		err = errors.New("Bad dates in request")
		http.Error(w, "Bad dates in request", http.StatusBadRequest)
		return
	}

	if getCheckType(chkReq.Host) == checkInvalid {
		err = errors.New("Host not acceptable")
		http.Error(w, "Host not acceptable", http.StatusBadRequest)
		return
	}

	err = checkHostExists(chkReq.Host)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Unknown host", http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return
}

const JsonChecksHandlerEndpoint string = "/api/checks"

func JsonChecksHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var chkReq ChecksRequest
	chkReq, err = GetChecksRequest(w, r)
	if err != nil {
		return
	}

	var cData []ChecksData
	cData, err = getChecksData(chkReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
