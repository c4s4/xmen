package main

import (
	"github.com/grandcat/zeroconf"
	"gopkg.in/yaml.v2"
	"os"
	"os/signal"
	"syscall"
	"io/ioutil"
	"fmt"
)

const DOMAIN = "local"

type Service struct {
	Desc    string
	Name    string
	Type    string
	Command []string
}

type Configuration struct {
	Services []Service
	Port     int
}

func LoadConfiguration(file string) (*Configuration, error) {
	var configuration Configuration
	source, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("ERROR reading configuration file: %v", err)
	}
	err = yaml.Unmarshal(source, &configuration)
	if err != nil {
		return nil, fmt.Errorf("ERROR parsing configuration file: %v", err)
	}
	return &configuration, nil
}

func RegisterServices(configuration *Configuration) ([]*zeroconf.Server, error) {
	var servers []*zeroconf.Server
	for _, service := range configuration.Services {
		fmt.Printf("Registering service %s\n", service.Name)
		server, err := zeroconf.Register(service.Name, service.Type, DOMAIN, configuration.Port, []string{""}, nil)
		if err != nil {
			panic(err)
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func WaitExit() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
	}
}

func Run(file string) error {
	configuration, err := LoadConfiguration(file)
	if err != nil {
		return fmt.Errorf("loading configuration: %v", err)
	}
	servers, err := RegisterServices(configuration)
	if err != nil {
		return fmt.Errorf("registering services: %v", err)
	}
	WaitExit()
	for _, server := range servers {
		server.Shutdown()
	}
	fmt.Println("Shutting down")
	return nil
}

func main() {
	if len(os.Args) != 2 {
		println("ERROR you must pass node configuration file on command line")
		os.Exit(1)
	}
	err := Run(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR %v\n", err)
		os.Exit(2)
	}
}
