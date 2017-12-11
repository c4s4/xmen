package main

import (
	"context"
	"github.com/grandcat/zeroconf"
	"sync"
	"fmt"
	"sort"
	"os"
	"os/signal"
	"syscall"
)

const (
	TYPE = "_publisher._tcp"
	DOMAIN = "local"
)

var publisherList PublisherList

type PublisherList struct {
	sync.RWMutex
	Publishers []*zeroconf.ServiceEntry
}

func (pl *PublisherList) AddPublisher(publisher *zeroconf.ServiceEntry) {
	pl.Lock()
	defer pl.Unlock()
	pl.Publishers = append(pl.Publishers, publisher)
	sort.Slice(pl.Publishers, func(i, j int) bool {
		return pl.Publishers[i].ServiceInstanceName() < pl.Publishers[j].ServiceInstanceName()
	})
}

func ProducePublishers() (chan *zeroconf.ServiceEntry, func(), error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize resolver: %v", err)
	}
	publisherChan := make(chan *zeroconf.ServiceEntry)
	context, cancel := context.WithCancel(context.Background())
	err = resolver.Browse(context, TYPE, DOMAIN, publisherChan)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to browse: %v", err)
	}
	return publisherChan, cancel, nil
}

func ConsumePublishers(publisherChan <-chan *zeroconf.ServiceEntry) {
	for publisher := range publisherChan {
		publisherList.AddPublisher(publisher)
		fmt.Printf("Added publisher %v\n", publisher)
	}
}

func WaitExit() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
	}
}

func main() {
	publisherChan, cancel, err := ProducePublishers()
	if err != nil {
		fmt.Printf("error browsing publishers: %v", err)
		os.Exit(1)
	}
	go ConsumePublishers(publisherChan)
	WaitExit()
	cancel()
	fmt.Println("Shutting down")
}
