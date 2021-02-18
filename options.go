package main

import (
	"encoding/json"
	"log"
	"os"
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
	}
	Chart struct {
		MaxRttScale     int64
		DynamicRttScale bool
	}
}

var Config = Configuration{}

func loadConfiguration(configPath string) {
	file, err := os.Open(configPath)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Config)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return
	}
	if Config.Checks.PingRetryCount < 1 {
		Config.Checks.PingRetryCount = 1
	}
	if Config.Chart.MaxRttScale <= 0 {
		Config.Chart.MaxRttScale = 200
	}
	if Config.DB.Type == "pg" {
		Config.DB.Type = "pq"
	}
}
