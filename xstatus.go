package main

import (
	"context"
	"log"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	TYPE = "_publisher._tcp"
	LOOKUP_DOMAIN = "local"
	WAIT_TIME = 10
)

func main() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println(entry)
		}
		log.Println("No more entries.")
	}(entries)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(WAIT_TIME))
	defer cancel()
	err = resolver.Browse(ctx, TYPE, LOOKUP_DOMAIN, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}
	<-ctx.Done()
	time.Sleep(1 * time.Second)
}
