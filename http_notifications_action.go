package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func PrepareEventAction(host string, rtt int64, checkTime time.Time, up bool, action string) string {
	action = strings.ReplaceAll(action, "{HOST}", url.QueryEscape(host))
	action = strings.ReplaceAll(action, "{TIME}", url.QueryEscape(checkTime.String()))
	action = strings.ReplaceAll(action, "{TIMESTAMP}", strconv.FormatInt(checkTime.Unix(), 10))
	action = strings.ReplaceAll(action, "{RTT}", strconv.FormatInt(rtt, 10))
	var rttstr time.Duration = time.Duration(rtt) * time.Nanosecond
	action = strings.ReplaceAll(action, "{RTTSTR}", url.QueryEscape(rttstr.String()))
	if up {
		action = strings.ReplaceAll(action, "{STATE}", "up")
		action = strings.ReplaceAll(action, "{UP}", "true")
		action = strings.ReplaceAll(action, "{DOWN}", "false")
	} else {
		action = strings.ReplaceAll(action, "{STATE}", "down")
		action = strings.ReplaceAll(action, "{UP}", "false")
		action = strings.ReplaceAll(action, "{DOWN}", "true")
	}
	return action
}

func EventHTTPNotify(host string, rtt int64, checkTime time.Time, up bool, action string) error {
	action = PrepareEventAction(host, rtt, checkTime, up, action)
	fmt.Println(action)

	client := &http.Client{
		Timeout: time.Duration(30) * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	var err error
	var req *http.Request
	req, err = http.NewRequest("GET", action, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	var resp *http.Response
	resp, err = client.Do(req)

	if resp != nil {
		err = resp.Body.Close()
		if err != nil {
			return err
		}
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 399) {
		return fmt.Errorf("Response status: %v", resp.StatusCode)
	}

	return nil
}
