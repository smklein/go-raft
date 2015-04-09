package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Server struct {
	Name    string
	Address string
	Port    int
}

type Config struct {
	Servers []Server
}

func LoadConfig(config *Config) bool {
	// Load config
	file, e := ioutil.ReadFile(os.Getenv("GOPATH") + "src/config.yaml")
	if e != nil {
		fmt.Println("Error: ", e)
		return false
	}
	e = yaml.Unmarshal(file, &config)
	if e != nil {
		fmt.Println("Error: ", e)
		return false
	}
	//fmt.Println("Config: ", config)
	return true
}

func (cfg *Config) GetServerNames() []string {
	serverNames := make([]string, len(cfg.Servers) - 1)
	i := 0
	for _, server := range(cfg.Servers) {
		if server.Name != "debugServer" {
			serverNames[i] = server.Name
			i++
		} 
	}
	return serverNames
}
