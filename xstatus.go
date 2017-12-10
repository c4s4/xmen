package main

import (
	"context"
	"time"
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
	WAIT_TIME = 500000
)

type PublisherList struct {
	sync.RWMutex
	Publishers []*zeroconf.ServiceEntry
}

func (pl *PublisherList) Update(entries chan *zeroconf.ServiceEntry) {
	pl.Lock()
	defer pl.Unlock()
	var publishers []*zeroconf.ServiceEntry
	sort.Slice(publishers, func(i, j int) bool {
		return publishers[i].ServiceInstanceName() < publishers[j].ServiceInstanceName()
	})
	pl.Publishers = publishers
}

func BrowsePublishers() (context.Context, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize resolver: %v", err)
	}
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			fmt.Println(entry)
		}
		fmt.Println("No more entries.")
	}(entries)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(WAIT_TIME))
	err = resolver.Browse(ctx, TYPE, DOMAIN, entries)
	if err != nil {
		return nil, fmt.Errorf("failed to browse: %v", err)
	}
	return ctx, nil
}

func WaitExit(ctx context.Context) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
	}
	ctx.Done()
	fmt.Println("Shutting down.")
}

func main() {
	ctx, err := BrowsePublishers()
	if err != nil {
		fmt.Printf("error browsing publishers: %v", err)
		os.Exit(1)
	}
	WaitExit(ctx)
}
