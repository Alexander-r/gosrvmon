package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

const indexTemplateDoc string = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>gosrvmon</title>
</head>

<body>

{{if .AllowSingleChecks}}
<h2>Check host</h2>
<form action="` + IndexTemplateHandlerRootEndpoint + `" method="get">
  <input name="host" type="text" value="{{.Host}}">
  <input type="submit" value="Check">
</form>
{{end}}

{{if .Check}}
<h2>Result</h2>
<p>Host is 
{{if .Up}}
<b>up</b>
{{else}}
<b>down</b>
{{end}}
 rtt: {{.Rtt}}
</p>
{{end}}

<h2>Web Endpoints</h2>
<ul>
  <li><a href="` + HostsTemplateHandlerEndpoint + `">` + HostsTemplateHandlerEndpoint + `</a></li>
  <li><a href="` + StateChangeParamsHandlerEndpoint + `">` + StateChangeParamsHandlerEndpoint + `</a></li>
  <li><a href="` + ChecksTemplateHandlerEndpoint + `">` + ChecksTemplateHandlerEndpoint + `</a></li>
  <li><a href="` + StateChangeParamsHandlerEndpoint + `">` + StateChangeParamsHandlerEndpoint + `</a></li>
  <li><a href="` + HostsViewTemplateHandlerEndpoint + `">` + HostsViewTemplateHandlerEndpoint + `</a></li>
</ul>

<h2>API Endpoints</h2>
<ul>
  <li><a href="` + JsonHostsHandlerEndpoint + `">` + JsonHostsHandlerEndpoint + `</a></li>
  <li><a href="` + JsonStateChangeParamsHandlerEndpoint + `">` + JsonStateChangeParamsHandlerEndpoint + `</a></li>
  <li><a href="` + JsonChecksHandlerEndpoint + `">` + JsonChecksHandlerEndpoint + `</a></li>
  <li><a href="` + JsonChecksLastHandlerEndpoint + `">` + JsonChecksLastHandlerEndpoint + `</a></li>
  <li><a href="` + JsonStateChangeParamsHandlerEndpoint + `">` + JsonStateChangeParamsHandlerEndpoint + `</a></li>
{{if .AllowSingleChecks}}
  <li><a href="` + JsonCheckHandlerEndpoint + `">` + JsonCheckHandlerEndpoint + `</a></li>
{{end}}
  <li><a href="` + ChecksChartEndpoint + `">` + ChecksChartEndpoint + `</a></li>
</ul>

</body>
</html>
`

const IndexTemplateHandlerRootEndpoint string = "/"
const IndexTemplateHandlerHtmlEndpoint string = "/index.html"

type IndexPageData struct {
	AllowSingleChecks bool
	Host              string
	Check             bool
	Up                bool
	Rtt               time.Duration
}

var indexTemplate = template.Must(template.New("Index Template").Parse(indexTemplateDoc))

func IndexTemplateHandler(w http.ResponseWriter, r *http.Request) {
	var rtt time.Duration
	var up bool = false
	var sCheck = false
	var err error
	var chkHost string
	if Config.Checks.AllowSingleChecks {
		chkHost = r.URL.Query().Get("host")
		if len(chkHost) > 0 {
			up, rtt, err = doSingleCheck(chkHost)
			if err == nil {
				sCheck = true
			}
		}
	}
	data := IndexPageData{
		AllowSingleChecks: Config.Checks.AllowSingleChecks,
		Host:              chkHost,
		Check:             sCheck,
		Up:                up,
		Rtt:               rtt,
	}

	err = indexTemplate.Execute(w, data)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}
