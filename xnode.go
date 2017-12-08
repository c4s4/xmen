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

const DOMAIN = "local."

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

func RegisterServices(configuration Configuration) ([]*zeroconf.Server, error) {
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

func WaitExit(servers []*zeroconf.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
	}
	for _, server := range servers {
		server.Shutdown()
	}
	fmt.Println("Shutting down.")
}

func main() {
	if len(os.Args) != 2 {
		println("You must pass node configuration file on command line")
		os.Exit(1)
	}
	file := os.Args[1]
	var configuration Configuration
    source, err := ioutil.ReadFile(file)
    if err != nil {
        fmt.Errorf("ERROR reading configuration file: %v", err)
		os.Exit(2)
    }
    err = yaml.Unmarshal(source, &configuration)
    if err != nil {
        fmt.Errorf("ERROR parsing configuration file: %v", err)
		os.Exit(2)
    }
    servers, err := RegisterServices(configuration)
    if err != nil {
    	fmt.Errorf("ERROR registering services: %v", err)
    	os.Exit(3)
	}
	WaitExit(servers)
}
