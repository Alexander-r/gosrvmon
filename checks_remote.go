package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func RedirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if Config.Listen.WebAuth.Enable {
		req.SetBasicAuth(Config.Listen.WebAuth.User, Config.Listen.WebAuth.Password)
	}
	return nil
}

func GetRemoteChecks(targetUrl string, checksReq ChecksRequest) (cData []ChecksData, err error) {
	cData = make([]ChecksData, 0)
	//TODO: separate config
	client := &http.Client{
		Timeout:       time.Duration(Config.Checks.Timeout) * time.Second,
		CheckRedirect: RedirectPolicyFunc,
	}

	var reqData []byte
	reqData, err = json.Marshal(checksReq)
	if err != nil {
		return cData, err
	}

	req, err := http.NewRequest("POST", targetUrl, bytes.NewBuffer(reqData))
	req.Header.Add("Content-Type", "application/json")

	if Config.Listen.WebAuth.Enable {
		req.SetBasicAuth(Config.Listen.WebAuth.User, Config.Listen.WebAuth.Password)
	}

	resp, err := client.Do(req)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return cData, err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return cData, fmt.Errorf("Response status: %v", resp.StatusCode)
	}

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return cData, err
	}
	err = json.Unmarshal(body, &cData)
	if err != nil {
		return cData, err
	}

	return cData, nil
}
