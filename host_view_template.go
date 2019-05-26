package main

import (
	"html/template"
	"log"
	"net/http"
)

const hostViewTemplateDoc string = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Host Statistics</title>
  <link rel="stylesheet" href="` + flatpickrCssEndpoint + `">
  <script src="` + flatpickrJsEndpoint + `"></script>
</head>

<body>

<form action="` + ChecksChartEndpoint + `" method="get">
  <input type="hidden" name="host" id="host" value="{{.}}">
  Start: <input name="start" class="flatpickr flatpickr-input active" type="text" placeholder="Select Date.." id="datetimestart" readonly="readonly"> End: <input name="end" class="flatpickr flatpickr-input active" type="text" placeholder="Select Date.." id="datetimeend" readonly="readonly">
  <input type="button" value="View" onclick="showChart()"><input type="submit" value="Open">
  <input type="button" value="&larr;" onclick="showPrev()"><input type="button" value="&rarr;" onclick="showNext()">
</form>



<div id="chart" style="margin-top: 25px;"><img alt="Chart" src="` + ChecksChartEndpoint + `?host={{.}}"></div>

<script>
	var timeNow = new Date();
	var timePrev = new Date(new Date - 86400000);
	const fpDateTimeEnd = flatpickr("#datetimeend", {
		enableTime: true,
		time_24hr: true,
		minuteIncrement: 1,
		defaultDate: timeNow,
		dateFormat: "U"
	});
	const fpDateTimeStart = flatpickr("#datetimestart", {
		enableTime: true,
		time_24hr: true,
		minuteIncrement: 1,
		defaultDate: timePrev,
		dateFormat: "U"
	});
	
	function showChart() {
		var chartHost = document.getElementById("host").value;
		var chartStart = document.getElementById("datetimestart").value;
		var chartEnd = document.getElementById("datetimeend").value;
		document.getElementById('chart').innerHTML = "<img alt=\"Chart\" src=\"` + ChecksChartEndpoint + `?host=" + chartHost + "&start=" + chartStart + "&end=" + chartEnd + "\">";
	}
	function showPrev() {
		var chartStart = parseInt(document.getElementById("datetimestart").value);
		var chartEnd = parseInt(document.getElementById("datetimeend").value);
		var chartDiff = chartEnd - chartStart;
		var chartNewStart = new Date((chartStart - chartDiff) * 1000);
		var chartNewEnd = new Date((chartEnd - chartDiff) * 1000);
		fpDateTimeStart.setDate(chartNewStart);
		fpDateTimeEnd.setDate(chartNewEnd);
		showChart();
	}
	function showNext() {
		var chartStart = parseInt(document.getElementById("datetimestart").value);
		var chartEnd = parseInt(document.getElementById("datetimeend").value);
		var chartDiff = chartEnd - chartStart;
		var chartNewEnd = new Date((chartEnd + chartDiff) * 1000);
		var chartNewStart = new Date((chartStart + chartDiff) * 1000);
		fpDateTimeEnd.setDate(chartNewEnd);
		fpDateTimeStart.setDate(chartNewStart);
		showChart();
	}
</script>

</body>
</html>
`

var hostViewTemplate = template.Must(template.New("Hosts View Template").Parse(hostViewTemplateDoc))

const HostsViewTemplateHandlerEndpoint string = "/web/view"

func HostsViewTemplateHandler(w http.ResponseWriter, r *http.Request) {
	var viewHost string
	viewHost = r.URL.Query().Get("host")
	err := hostViewTemplate.Execute(w, viewHost)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}
}
