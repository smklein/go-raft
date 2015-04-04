package server

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Server struct {
	Name string
	Address string
	Port int
}

type Config struct {
	Servers []Server
}

func LoadConfig(config *Config) bool {
	// Load config
	file, e := ioutil.ReadFile("../src/config.yaml")
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