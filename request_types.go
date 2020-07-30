package main

import (
	"github.com/docker/go-connections/nat"
)

type ContainerInstance struct {
	ServiceName string
	ImageName   string
	Ports       nat.PortSet
}
