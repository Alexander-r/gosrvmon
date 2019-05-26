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
<form action="` + JsonHostsHandlerEndpoint + `" method="get">
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
	<td><a href="` + JsonHostsHandlerEndpoint + `?action=del&host={{.}}">del</a></td>
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
	hostsList, err := getHostsList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := HostsPageData{
		Created: false,
		Deleted: false,
		Hosts:   hostsList,
	}
	action := r.URL.Query().Get("action")
	if action == "created" {
		data.Created = true
	}
	if action == "deleted" {
		data.Deleted = true
	}
	err = hostsTemplate.Execute(w, data)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}
