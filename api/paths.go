package api

import (
	"fmt"
	"strconv"

	utils "github.com/bruno-anjos/solution-utils"
)

// Paths
const (
	PrefixPath = "/scheduler"

	InstancesPath = "/instances"
	InstancePath  = "/instances/%s"
)

const (
	Port = 50001
)

var (
	DefaultHostPort = utils.LocalhostAddr + ":" + strconv.Itoa(Port)
)

func GetInstancesPath() string {
	return PrefixPath + InstancesPath
}

func GetInstancePath(instanceId string) string {
	return PrefixPath + fmt.Sprintf(InstancePath, instanceId)
}
