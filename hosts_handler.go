package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

func AddHost(newHost string) error {
	if getCheckType(newHost) == checkInvalid {
		return errors.New("Host not acceptable")
	}
	var err error
	err = checkHostExists(newHost)
	if err != sql.ErrNoRows {
		if err == nil {
			return errors.New("Bad request")
		}
		return err
	}
	err = addHost(newHost)
	if err != nil {
		return err
	}
	return nil
}

func DeleteHost(newHost string) error {
	if getCheckType(newHost) == checkInvalid {
		return errors.New("Host not acceptable")
	}
	var err error
	err = checkHostExists(newHost)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Host not in DB")
		} else {
			return err
		}
	}
	err = deleteHost(newHost)
	if err != nil {
		return err
	}
	return nil
}

const JsonHostsHandlerEndpoint string = "/api/hosts"

func JsonHostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		action := r.URL.Query().Get("action")
		if len(action) == 0 {
			hosts, err := getHostsList()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			jsonData, err := json.Marshal(hosts)
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

		if action == "add" {
			newHost := r.URL.Query().Get("host")
			if len(newHost) <= 0 {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			err := AddHost(newHost)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Redirect(w, r, HostsTemplateHandlerEndpoint+"?action=created", http.StatusSeeOther)
			return
		}

		if action == "del" {
			newHost := r.URL.Query().Get("host")
			if len(newHost) <= 0 {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			err := DeleteHost(newHost)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Redirect(w, r, HostsTemplateHandlerEndpoint+"?action=deleted", http.StatusSeeOther)
			return
		}

		http.Error(w, "Bad request", http.StatusBadRequest)
		return

	case http.MethodPost:
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
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		var newHost string
		err = json.Unmarshal(body, &newHost)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = AddHost(newHost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	case http.MethodDelete:
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
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		var newHost string
		err = json.Unmarshal(body, &newHost)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = DeleteHost(newHost)
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
