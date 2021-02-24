package main

import (
	"encoding/json"
	"os"
	"strings"
)

type Configuration struct {
	DB struct {
		Type     string
		Host     string
		Port     string
		User     string
		Password string
		Database string
	}
	Listen struct {
		Address      string
		Port         string
		ReadTimeout  int64
		WriteTimeout int64
		WebAuth      struct {
			Enable   bool
			User     string
			Password string
		}
	}
	Checks struct {
		Timeout           int64
		Interval          int64
		PingRetryCount    uint32
		HTTPMethod        string
		PerformChecks     bool
		UseRemoteChecks   bool
		RemoteChecksURLs  []string
		AllowSingleChecks bool
		Retention         int64
	}
	Chart struct {
		MaxRttScale     int64
		DynamicRttScale bool
	}
}

var Config = Configuration{}

func loadConfiguration() error {
	Config.Checks.PerformChecks = true

	if *configPath != "" {
		file, err := os.Open(*configPath)
		if err != nil {
			return err
		}
		defer file.Close()
		decoder := json.NewDecoder(file)

		err = decoder.Decode(&Config)
		if err != nil {
			return err
		}
	}
	if *configStr != "" {
		cfgBuf := strings.NewReader(*configStr)
		decoder := json.NewDecoder(cfgBuf)

		err := decoder.Decode(&Config)
		if err != nil {
			return err
		}
	}

	if Config.Listen.Port == "" {
		Config.Listen.Port = "8000"
	}
	if Config.Listen.ReadTimeout == 0 {
		Config.Listen.ReadTimeout = 30
	}
	if Config.Listen.WriteTimeout == 0 {
		Config.Listen.WriteTimeout = 60
	}
	if Config.Checks.Timeout == 0 {
		Config.Checks.Timeout = 10
	}
	if Config.Checks.Interval == 0 {
		Config.Checks.Interval = 60
	}
	if Config.Checks.PingRetryCount < 1 {
		Config.Checks.PingRetryCount = 1
	}
	if Config.Checks.HTTPMethod == "" {
		Config.Checks.HTTPMethod = "GET"
	}
	if Config.Chart.MaxRttScale <= 0 {
		Config.Chart.MaxRttScale = 200
	}
	if Config.DB.Type == "pg" {
		Config.DB.Type = "pq"
	}
	if Config.DB.Type == "bbolt" {
		Config.DB.Type = "bolt"
	}
	if Config.DB.Database == "" {
		Config.DB.Type = "ql"
	}
	return nil
}
