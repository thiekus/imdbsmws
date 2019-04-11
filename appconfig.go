package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gorilla/securecookie"
	"io/ioutil"
	"os"
)

type ConfigData struct {
	Username string		`json:"username"`
	Password string		`json:"password"`
	SessionKey string	`json:"sessionKey"`
	ListeningPort int	`json:"listeningPort"`
}

const ConfigDefaultUsername = "admin"
const ConfigDefaultPassword = "admin"
const ConfigDefaultPort = 33666
const ConfigFilename = "config.json"

// Get configuration from config file
func getConfigData() ConfigData {
	log:= newLog()
	configPath:= "./" + ConfigFilename
	cfg:= ConfigData{}
	if !isFileExists(configPath) {
		cfg.Username = ConfigDefaultUsername
		cfg.Password = fmt.Sprintf("%x", sha256.Sum256([]byte(ConfigDefaultPassword)))
		cfg.SessionKey = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%x", securecookie.GenerateRandomKey(32)))))
		cfg.ListeningPort = ConfigDefaultPort
		log.Printf("Creating new admin username=%s; password=%s", ConfigDefaultUsername, ConfigDefaultPassword)
		saveConfigData(cfg)
	}
	if jsonData, err:= ioutil.ReadFile(configPath); err == nil {
		if err = json.Unmarshal(jsonData, &cfg); err == nil {
			log.Print("Config data loaded...")
		} else {
			log.Fatal("Cannot parse configuration: "+err.Error())
		}
	} else {
		log.Fatal("Cannot read configuration: "+err.Error())
	}
	return cfg
}

// Save current configuration to config file
func saveConfigData(config ConfigData) {
	configPath:= "./" + ConfigFilename
	if jsonData, err:= json.Marshal(config); err == nil {
		if err = ioutil.WriteFile(configPath, jsonData, os.ModePerm); err == nil {
			log:= newLog()
			log.Print("Config data have been saved!")
		}
	}
}
