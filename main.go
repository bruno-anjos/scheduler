package main

import (
	"github.com/bruno-anjos/scheduler/api"
	utils "github.com/bruno-anjos/solution-utils"
)

const (
	serviceName = "SCHEDULER"
)

func main() {
	utils.StartServer(serviceName, api.DefaultHostPort, api.Port, api.PrefixPath, routes)
}
