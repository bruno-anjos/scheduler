package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bruno-anjos/scheduler/api"
	utils "github.com/bruno-anjos/solution-utils"
)

const (
	serviceName = "SCHEDULER"
)

func main() {
	setupCloseHandler()
	utils.StartServer(serviceName, api.DefaultHostPort, api.Port, api.PrefixPath, routes)
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		deleteAllInstances()
		os.Exit(0)
	}()
}
