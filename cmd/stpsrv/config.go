package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
)

type GlobalConfig struct {
	AuthKey     string `json:"authKey"`
	ListentAddr string `json:"listenAddr"`
	SSHAddr     string `json:"sshAddr"`
	SSHUser     string `json:"sshUser"`
	SSHRSAPath  string `json:"sshRsaPath"`
	PortRange   string `json:"portRange"`
}

var config = &GlobalConfig{}

func Config() *GlobalConfig {
	return config
}

func ParseConfig(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil && err != io.EOF {
		log.Fatalln("load config fail, error", err.Error())
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		log.Fatalln("load config fail, error", err.Error())
	}
}
