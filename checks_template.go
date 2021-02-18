package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

const checksTemplateDoc string = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Checks</title>
</head>

<body>

<h2>Checks</h2>
<table id="checks">
{{range .}}
  <tr>
	<td>{{.Timestamp}}</td>
	<td>{{.Rtt}}</td>
    <td>{{.Up}}</td>
  </tr>
{{end}}
</table>

</body>
</html>
`

type ChecksPageData struct {
	Created bool
	Deleted bool
	Hosts   []string
}

var checksTemplate = template.Must(template.New("Checks Template").Parse(checksTemplateDoc))

const ChecksTemplateHandlerEndpoint string = "/web/checks"

func ChecksTemplateHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var chkReq ChecksRequest
	chkReq, err = GetChecksRequest(w, r)
	if err != nil {
		return
	}

	var data []ChecksData
	data, err = MonData.GetChecksData(chkReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//RemoteChecks
	dataM := make(map[time.Time]ChecksData)
	for _, d := range data {
		dataM[d.Timestamp.UTC()] = d
	}

	if Config.Checks.UseRemoteChecks {
		for _, CheckURL := range Config.Checks.RemoteChecksURLs {
			var RemoteData []ChecksData
			RemoteData, err = GetRemoteChecks(CheckURL, chkReq)
			if err != nil {
				//TODO: print error and ignore Response status 400
				//log.Printf("[ERROR] %v", err)
				continue
			}
			for _, dr := range RemoteData {
				dl, ok := dataM[dr.Timestamp]
				if ok {
					if dr.Up {
						if !dl.Up {
							dataM[dr.Timestamp] = dr
						} else if dr.Rtt < dl.Rtt {
							dataM[dr.Timestamp] = dr
						}
					}
				} else {
					dataM[dr.Timestamp] = dr
				}
			}
		}
	}
	//RemoteChecks end

	err = checksTemplate.Execute(w, data)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}

const ChecksChartEndpoint string = "/web/checks/svg"

func checksChart(w http.ResponseWriter, r *http.Request) {
	var err error
	var chkReq ChecksRequest
	chkReq, err = GetChecksRequest(w, r)
	if err != nil {
		return
	}

	var data []ChecksData
	data, err = MonData.GetChecksData(chkReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//RemoteChecks
	dt := time.Duration(Config.Checks.Interval) * time.Second
	dataM := make(map[time.Time]ChecksData)
	var maxRtt int64 = 0
	for _, d := range data {
		dataM[d.Timestamp.Truncate(dt).UTC()] = d
		if d.Rtt > maxRtt {
			maxRtt = d.Rtt
		}
	}

	if Config.Checks.UseRemoteChecks {
		for _, CheckURL := range Config.Checks.RemoteChecksURLs {
			var RemoteData []ChecksData
			RemoteData, err = GetRemoteChecks(CheckURL, chkReq)
			if err != nil {
				//TODO: print error and ignore Response status 400
				//log.Printf("[ERROR] %v", err)
				continue
			}
			for _, dr := range RemoteData {
				ts := dr.Timestamp.Truncate(dt).UTC()
				dr.Timestamp = ts
				dl, ok := dataM[ts]
				if ok {
					if dr.Up {
						if !dl.Up {
							dataM[ts] = dr
							if dr.Rtt > maxRtt {
								maxRtt = dr.Rtt
							}
						} else if dr.Rtt < dl.Rtt {
							dataM[ts] = dr
							if dr.Rtt > maxRtt {
								maxRtt = dr.Rtt
							}
						}
					}
				} else {
					dataM[ts] = dr
				}
			}
		}
	}
	//RemoteChecks end

	var chartMaxRtt int64 = Config.Chart.MaxRttScale

	maxRtt = (maxRtt / 1000000) + 1
	if Config.Chart.DynamicRttScale {
		dynamicMaxRtt := (maxRtt / 100) * 100
		if maxRtt%100 > 0 {
			dynamicMaxRtt += 100
		}
		if dynamicMaxRtt < Config.Chart.MaxRttScale && dynamicMaxRtt > 0 {
			chartMaxRtt = dynamicMaxRtt
		}
	}

	chart := getChart(1280, 720, chartMaxRtt, chkReq, &dataM)

	w.Header().Set("Content-Type", "image/svg+xml")
	_, err = w.Write([]byte(chart))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
