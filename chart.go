package main

import (
	"strconv"
	"strings"
	"time"
)

type ChartState int32

const (
	chartStateUp      ChartState = 0
	chartStateDown    ChartState = 1
	chartStateUnknown ChartState = 2
	chartStateFirst   ChartState = 3
)

type ChartData struct {
	Timestamp time.Time
	Rtt       int64
	Up        ChartState
}

func getChart(width int64, height int64, chkReq ChecksRequest, dataM *map[time.Time]ChecksData) string {
	dt := time.Duration(Config.Checks.Interval) * time.Second
	if chkReq.Start.Truncate(dt).Equal(chkReq.End.Truncate(dt)) || chkReq.Start.Truncate(dt).Add(dt).After(chkReq.End.Truncate(dt)) {
		//TODO: error
		return ""
	}

	var fontSize int64 = 12
	var chart strings.Builder
	var xOffset = fontSize * 3
	var yOffsetBottom = height - fontSize*3
	var yOffsetTop = fontSize * 3
	chart.WriteString("<svg width=\"")
	chart.WriteString(strconv.FormatInt(width, 10))
	chart.WriteString("\" height=\"")
	chart.WriteString(strconv.FormatInt(height, 10))
	chart.WriteString("\" viewBox=\"0 0 ")
	chart.WriteString(strconv.FormatInt(width, 10))
	chart.WriteString(" ")
	chart.WriteString(strconv.FormatInt(height, 10))
	chart.WriteString("\" shape-rendering=\"crispEdges\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\">")
	chart.WriteString(`
<style>
/* <![CDATA[ */
text {
  font-family: 'Roboto Medium',sans-serif";
  stroke-width: 0;
  stroke: none;
  fill: rgba(51,51,51,1.0);
  font-size: `)
	chart.WriteString(strconv.FormatInt(fontSize, 10))
	chart.WriteString(`px;
}
text.stat {
  font-family: 'Roboto Medium',sans-serif";
  stroke-width: 0;
  stroke: none;
  fill: rgba(51,51,51,1.0);
  font-size: `)
	chart.WriteString(strconv.FormatInt(fontSize*2, 10))
	chart.WriteString(`px;
  text-anchor:middle;
}
path {
  stroke-width: 1;
  stroke: rgba(51,51,51,1.0);
  fill: none;
}
rect.up {
  stroke-width: 0;
  stroke: none;
  fill: rgba(102,255,102,1.0);
}
rect.down {
  stroke-width: 0;
  stroke: none;
  fill: rgba(255,102,102,1.0);
}
rect.na {
  stroke-width: 0;
  stroke: none;
  fill: rgba(102,102,102,1.0);
}
rect.i {
  stroke-width: 0;
  stroke: none;
  fill: rgba(102,102,255,1.0);
}
/* ]]> */
</style>
`)
	//ms
	chart.WriteString("<text x=\"0\" y=\"")
	chart.WriteString(strconv.FormatInt(fontSize, 10))
	chart.WriteString("\">ms</text>\n")
	//T
	chart.WriteString("<text x=\"0\" y=\"")
	chart.WriteString(strconv.FormatInt(height-fontSize, 10))
	chart.WriteString("\">T</text>\n")
	//axis lines
	chart.WriteString("<path  d=\"M ")
	chart.WriteString(strconv.FormatInt(xOffset, 10))
	chart.WriteString(" 0 L ")
	chart.WriteString(strconv.FormatInt(xOffset, 10))
	chart.WriteString(" ")
	chart.WriteString(strconv.FormatInt(yOffsetBottom, 10))
	chart.WriteString(" L ")
	chart.WriteString(strconv.FormatInt(width, 10))
	chart.WriteString(" ")
	chart.WriteString(strconv.FormatInt(yOffsetBottom, 10))
	chart.WriteString("\"/>\n")
	//x axis timeline
	var numberOfStepsX int64 = 24
	totalT := chkReq.End.Truncate(dt).Sub(chkReq.Start.Truncate(dt))
	stepT := totalT.Nanoseconds() / numberOfStepsX
	stepPx := float64(width-xOffset) / float64(numberOfStepsX)
	for i := int64(0); i < numberOfStepsX; i++ {
		chart.WriteString("<path  d=\"M ")
		chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepPx*float64(i), 'f', -1, 64))
		chart.WriteString(" ")
		chart.WriteString(strconv.FormatInt(yOffsetBottom, 10))
		chart.WriteString(" L ")
		chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepPx*float64(i), 'f', -1, 64))
		chart.WriteString(" ")
		chart.WriteString(strconv.FormatInt(yOffsetBottom+5, 10))
		chart.WriteString("\"/>\n")
		chart.WriteString("<text x=\"")
		chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepPx*float64(i), 'f', -1, 64))
		chart.WriteString("\" y=\"")
		chart.WriteString(strconv.FormatFloat(float64(height)-float64(fontSize)*1.6, 'f', -1, 64))
		chart.WriteString("\">")
		chart.WriteString(chkReq.Start.Truncate(dt).Add(time.Duration(stepT*i) * time.Nanosecond).Format("06-01-02"))
		chart.WriteString("</text>\n")
		chart.WriteString("<text x=\"")
		chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepPx*float64(i), 'f', -1, 64))
		chart.WriteString("\" y=\"")
		chart.WriteString(strconv.FormatFloat(float64(height)-float64(fontSize)*0.6, 'f', -1, 64))
		chart.WriteString("\">")
		chart.WriteString(chkReq.Start.Truncate(dt).Add(time.Duration(stepT*i) * time.Nanosecond).Format("15:04:05"))
		chart.WriteString("</text>\n")
	}
	//TODO: replace 200 ms with calculated value for scale
	//Config.Checks.Timeout
	//y axis milliseconds
	var numberOfStepsY int64 = 20
	scaleTimeout := time.Millisecond * 200
	stepTy := scaleTimeout.Nanoseconds() / 1000000 / numberOfStepsY
	stepPxy := float64(yOffsetBottom-yOffsetTop) / float64(numberOfStepsY)
	for i := int64(0); i <= numberOfStepsY; i++ {
		chart.WriteString("<path  d=\"M ")
		chart.WriteString(strconv.FormatInt(xOffset, 10))
		chart.WriteString(" ")
		chart.WriteString(strconv.FormatFloat(float64(yOffsetBottom)-stepPxy*float64(i), 'f', -1, 64))
		chart.WriteString(" L ")
		chart.WriteString(strconv.FormatInt(xOffset-5, 10))
		chart.WriteString(" ")
		chart.WriteString(strconv.FormatFloat(float64(yOffsetBottom)-stepPxy*float64(i), 'f', -1, 64))
		chart.WriteString("\"/>\n")
		chart.WriteString("<text x=\"0\" y=\"")
		chart.WriteString(strconv.FormatFloat(float64(yOffsetBottom)-stepPxy*float64(i), 'f', -1, 64))
		chart.WriteString("\">")
		chart.WriteString(strconv.FormatInt(stepTy*i, 10))
		chart.WriteString("</text>\n")
	}
	//check results
	stepIx := float64(width-xOffset) / (chkReq.End.Truncate(dt).Sub(chkReq.Start.Truncate(dt)).Seconds() / float64(Config.Checks.Interval))
	stepIy := float64(yOffsetBottom-yOffsetTop) / float64(scaleTimeout.Nanoseconds())
	var i int64 = 0
	var statUp int64 = 0
	var statDown int64 = 0
	var statNa int64 = 0
	var maxHeight = float64(yOffsetBottom - yOffsetTop)
	prevState := chartStateFirst
	var prevStateCount int64 = 0
	var chartRtt strings.Builder
	for t := chkReq.Start.Truncate(dt).Add(dt); t.Before(chkReq.End.Truncate(dt)) || t.Equal(chkReq.End.Truncate(dt)); t = t.Add(dt) {
		var curState ChartState
		d, ok := (*dataM)[t.UTC()]
		if ok {
			if d.Up {
				curState = chartStateUp
				chartRtt.WriteString("<rect x=\"")
				chartRtt.WriteString(strconv.FormatFloat(float64(xOffset)+stepIx*float64(i), 'f', -1, 64))
				chartRtt.WriteString("\" y=\"")
				var curHeight = stepIy * (float64(d.Rtt))
				if curHeight > maxHeight {
					curHeight = maxHeight
				}
				chartRtt.WriteString(strconv.FormatFloat(float64(height-yOffsetTop)-curHeight, 'f', -1, 64))
				chartRtt.WriteString("\" width=\"")
				chartRtt.WriteString(strconv.FormatFloat(stepIx, 'f', -1, 64))
				chartRtt.WriteString("\" height=\"")
				chartRtt.WriteString(strconv.FormatFloat(curHeight, 'f', -1, 64))
				chartRtt.WriteString("\" class=\"i\"/>\n")
				statUp++
			} else {
				curState = chartStateDown
				statDown++
			}
		} else {
			curState = chartStateUnknown
			statNa++
		}
		if prevState != curState {
			switch prevState {
			case chartStateUp:
				chart.WriteString("<rect x=\"")
				chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepIx*float64(i-prevStateCount), 'f', -1, 64))
				chart.WriteString("\" y=\"")
				chart.WriteString(strconv.FormatInt(yOffsetTop, 10))
				chart.WriteString("\" width=\"")
				chart.WriteString(strconv.FormatFloat(stepIx*float64(prevStateCount), 'f', -1, 64))
				chart.WriteString("\" height=\"")
				chart.WriteString(strconv.FormatInt(yOffsetBottom-yOffsetTop, 10))
				chart.WriteString("\" class=\"up\"/>\n")
			case chartStateDown:
				chart.WriteString("<rect x=\"")
				chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepIx*float64(i-prevStateCount), 'f', -1, 64))
				chart.WriteString("\" y=\"")
				chart.WriteString(strconv.FormatInt(yOffsetTop, 10))
				chart.WriteString("\" width=\"")
				chart.WriteString(strconv.FormatFloat(stepIx*float64(prevStateCount), 'f', -1, 64))
				chart.WriteString("\" height=\"")
				chart.WriteString(strconv.FormatInt(yOffsetBottom-yOffsetTop, 10))
				chart.WriteString("\" class=\"down\"/>\n")
			case chartStateUnknown:
				chart.WriteString("<rect x=\"")
				chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepIx*float64(i-prevStateCount), 'f', -1, 64))
				chart.WriteString("\" y=\"")
				chart.WriteString(strconv.FormatInt(yOffsetTop, 10))
				chart.WriteString("\" width=\"")
				chart.WriteString(strconv.FormatFloat(stepIx*float64(prevStateCount), 'f', -1, 64))
				chart.WriteString("\" height=\"")
				chart.WriteString(strconv.FormatInt(yOffsetBottom-yOffsetTop, 10))
				chart.WriteString("\" class=\"na\"/>\n")
			}
			prevState = curState
			prevStateCount = 1
		} else {
			prevStateCount++
		}
		i++
	}
	//TODO: avoid copy
	switch prevState {
	case chartStateUp:
		chart.WriteString("<rect x=\"")
		chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepIx*float64(i-prevStateCount), 'f', -1, 64))
		chart.WriteString("\" y=\"")
		chart.WriteString(strconv.FormatInt(yOffsetTop, 10))
		chart.WriteString("\" width=\"")
		chart.WriteString(strconv.FormatFloat(stepIx*float64(prevStateCount), 'f', -1, 64))
		chart.WriteString("\" height=\"")
		chart.WriteString(strconv.FormatInt(yOffsetBottom-yOffsetTop, 10))
		chart.WriteString("\" class=\"up\"/>\n")
	case chartStateDown:
		chart.WriteString("<rect x=\"")
		chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepIx*float64(i-prevStateCount), 'f', -1, 64))
		chart.WriteString("\" y=\"")
		chart.WriteString(strconv.FormatInt(yOffsetTop, 10))
		chart.WriteString("\" width=\"")
		chart.WriteString(strconv.FormatFloat(stepIx*float64(prevStateCount), 'f', -1, 64))
		chart.WriteString("\" height=\"")
		chart.WriteString(strconv.FormatInt(yOffsetBottom-yOffsetTop, 10))
		chart.WriteString("\" class=\"down\"/>\n")
	case chartStateUnknown:
		chart.WriteString("<rect x=\"")
		chart.WriteString(strconv.FormatFloat(float64(xOffset)+stepIx*float64(i-prevStateCount), 'f', -1, 64))
		chart.WriteString("\" y=\"")
		chart.WriteString(strconv.FormatInt(yOffsetTop, 10))
		chart.WriteString("\" width=\"")
		chart.WriteString(strconv.FormatFloat(stepIx*float64(prevStateCount), 'f', -1, 64))
		chart.WriteString("\" height=\"")
		chart.WriteString(strconv.FormatInt(yOffsetBottom-yOffsetTop, 10))
		chart.WriteString("\" class=\"na\"/>\n")
	}
	chart.WriteString(chartRtt.String())

	chart.WriteString("<text class=\"stat\" x=\"50%\" y=\"")
	chart.WriteString(strconv.FormatInt(fontSize*2, 10))
	chart.WriteString("\">")
	if true {
		chart.WriteString("Host: ")
		chart.WriteString(chkReq.Host)
		chart.WriteString(" ")
	}
	chart.WriteString("Up ")
	chart.WriteString(strconv.FormatFloat(float64(100*statUp)/float64(i), 'f', 2, 64))
	chart.WriteString("% Down ")
	chart.WriteString(strconv.FormatFloat(float64(100*statDown)/float64(i), 'f', 2, 64))
	chart.WriteString("% Unknown ")
	chart.WriteString(strconv.FormatFloat(float64(100*statNa)/float64(i), 'f', 2, 64))
	chart.WriteString("%</text>\n")
	chart.WriteString("</svg>\n")
	return chart.String()
}
