package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

const stateChangeParamsTemplateDoc string = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>State Change Params</title>
  <style>
    td {padding-right: 1em;}
  </style>
</head>

<body>

<h2>Add</h2>
<form action="` + StateChangeParamsHandlerEndpoint + `" method="post">
  <input type="hidden" name="action" value="add">
  <p>Host:</p>
  <p><input name="host" type="text"></p>
  <p>Change threshold:</p>
  <p><input name="threshold" type="number"></p>
  <p>Action:</p>
  <p><input name="state_action" type="text"></p>
  <p><input type="submit" value="Add"></p>
</form>

{{if .Created}}
<h2>Created</h2>
{{end}}

{{if .Deleted}}
<h2>Deleted</h2>
{{end}}

<h2>State Change Params</h2>
<table id="hosts">
  <tr>
	<th>Host</a></th>
	<th>Threshold</th>
	<th>Action</th>
	<th>Delete</th>
  </tr>
{{range .Params}}
  <tr>
	<td><a href="` + HostsViewTemplateHandlerEndpoint + `?host={{.Host}}">{{.Host}}</a></td>
	<td>{{.ChangeThreshold}}</td>
	<td>{{.Action}}</td>
	<td><form action="` + StateChangeParamsHandlerEndpoint + `" method="post">
	  <input type="hidden" name="action" value="del">
	  <input type="hidden" name="host" value="{{.Host}}">
	  <input type="submit" value="Delete">
	</form></td>
  </tr>
{{end}}
</table>

</body>
</html>
`

type StateChangeParamsPageData struct {
	Created bool
	Deleted bool
	Params  []StateChangeParams
}

var stateChangeParamsTemplate = template.Must(template.New("State Change Params Template").Parse(stateChangeParamsTemplateDoc))

const StateChangeParamsHandlerEndpoint string = "/web/state_change_params"

func StateChangeParamsTemplateHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	data := StateChangeParamsPageData{
		Created: false,
		Deleted: false,
		Params:  nil,
	}
	action := r.URL.Query().Get("action")
	if action == "created" {
		data.Created = true
	}
	if action == "deleted" {
		data.Deleted = true
	}

	if r.Method == http.MethodGet {
		action := r.URL.Query().Get("action")
		if action == "created" {
			data.Created = true
		}
		if action == "deleted" {
			data.Deleted = true
		}
	}

	if r.Method == http.MethodPost {
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

		action := r.PostFormValue("action")
		newHost := r.PostFormValue("host")
		newThreshold := r.PostFormValue("threshold")
		newAction := r.PostFormValue("state_action")
		if len(newHost) <= 0 {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if action != "add" && action != "del" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if action == "add" {
			if len(newThreshold) <= 0 {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			var checkThreshold int64
			checkThreshold, err = strconv.ParseInt(newThreshold, 10, 64)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			err = addHostStateChangeParams(newHost, checkThreshold, newAction)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data.Created = true
		}

		if action == "del" {
			err := deleteHostStateChangeParams(newHost)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data.Deleted = true
		}
	}

	var p []StateChangeParams
	p, err = getHostStateChangeParamsList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Params = p

	err = stateChangeParamsTemplate.Execute(w, data)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}
