package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type AppConfig struct {
	CapiUrl    string
	CapiApiKey string
	UsageUrl   string
	GridApiKey string
	GobbyFile  string
	FromDate   string
	ToDate     string
}

func LoadConfig() *AppConfig {
	var c *AppConfig = &AppConfig{}

	configBytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(configBytes, c)
	if err != nil {
		log.Fatal(err)
	}

	return c
}
