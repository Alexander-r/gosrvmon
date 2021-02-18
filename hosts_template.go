package main

import (
	"html/template"
	"log"
	"net/http"
)

const hostsTemplateDoc string = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Hosts</title>
  <style>
    td {padding-right: 1em;}
  </style>
</head>

<body>

<h2>Add host</h2>
<form action="` + HostsTemplateHandlerEndpoint + `" method="post">
  <input type="hidden" name="action" value="add">
  <input name="host" type="text">
  <input type="submit" value="Add">
</form>

{{if .Created}}
<h2>Host created</h2>
{{end}}

{{if .Deleted}}
<h2>Host deleted</h2>
{{end}}

<h2>Hosts</h2>
<table id="hosts">
{{range .Hosts}}
  <tr>
	<td><a href="` + HostsViewTemplateHandlerEndpoint + `?host={{.}}">{{.}}</a></td>
	<td><a href="` + ChecksChartEndpoint + `?host={{.}}">Last day</a></td>
	<td><form action="` + HostsTemplateHandlerEndpoint + `" method="post">
	  <input type="hidden" name="action" value="del">
	  <input type="hidden" name="host" value="{{.}}">
	  <input type="submit" value="Delete">
	</form></td>
  </tr>
{{end}}
</table>

</body>
</html>
`

type HostsPageData struct {
	Created bool
	Deleted bool
	Hosts   []string
}

var hostsTemplate = template.Must(template.New("Hosts Template").Parse(hostsTemplateDoc))

const HostsTemplateHandlerEndpoint string = "/web/hosts"

func HostsTemplateHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	data := HostsPageData{
		Created: false,
		Deleted: false,
		Hosts:   nil,
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
		if len(newHost) <= 0 {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if action != "add" && action != "del" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if action == "add" {
			err := AddHost(newHost)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data.Created = true
		}

		if action == "del" {
			err := DeleteHost(newHost)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			data.Deleted = true
		}
	}

	var hostsList []string
	hostsList, err = MonData.GetHostsList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data.Hosts = hostsList

	err = hostsTemplate.Execute(w, data)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}
