package api

import (
	"github.com/docker/go-connections/nat"
)

type InstanceDTO struct {
	Static          bool
	PortTranslation nat.PortMap `json:"port_translation"`
	Local           bool
}
