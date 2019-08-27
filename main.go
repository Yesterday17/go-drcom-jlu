package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Yesterday17/go-drcom-jlu/config"
	"github.com/Yesterday17/go-drcom-jlu/drcom"
)

// return code list
// -10 failed to parse config file

func main() {
	// cfg, err := config.ReadConfig("./config.json")
	cfg, err := config.ReadConfig("/etc/go-drcom-jlu/config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(-10)
	}

	client := drcom.New(&cfg)
	client.Start()

	// handle signal
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-sigc
		log.Printf("[GDJ] Exiting with signal %s", s.String())
		client.Close()
		return
	}
}
