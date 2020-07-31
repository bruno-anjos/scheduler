package api

import (
	"github.com/docker/go-connections/nat"
)

type ContainerInstanceDTO struct {
	ServiceName string `json:"service_name"`
	ImageName   string `json:"image_name"`
	Ports       nat.PortSet
}
