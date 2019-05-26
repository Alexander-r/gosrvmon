package main

import (
	"encoding/json"
	"net/http"
)

const JsonChecksLastHandlerEndpoint string = "/api/checks/last"

func JsonChecksLastHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var chkHost string
	chkHost, err = GetCheckRequest(w, r)
	if err != nil {
		return
	}

	var cData ChecksData

	cData, err = getLastCheckData(chkHost)
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
