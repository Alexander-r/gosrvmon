package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const JsonStateChangeParamsHandlerEndpoint string = "/api/notifications_params"

func JsonStateChangeParamsHandler(w http.ResponseWriter, r *http.Request) {
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

	switch r.Method {
	case http.MethodGet:
		p, err := MonData.GetHostStateChangeParamsList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(p)
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

	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		var newParams StateChangeParams
		err = json.Unmarshal(body, &newParams)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = MonData.AddHostStateChangeParams(newParams.Host, newParams.ChangeThreshold, newParams.Action)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return

	case http.MethodDelete:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		var newHost string = string(body)

		err = MonData.DeleteHostStateChangeParams(newHost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
