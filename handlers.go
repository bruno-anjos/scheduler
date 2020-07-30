package main

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	archimedes "github.com/bruno-anjos/archimedes/api"
	generic_utils "github.com/bruno-anjos/solution-utils"
	"github.com/bruno-anjos/solution-utils/http_utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	localhost         = "127.0.0.1"
	defaultInterface  = "0.0.0.0"
	httpClientTimeout = 10
)

var (
	dockerClient *client.Client
	httpClient   *http.Client
)

func init() {
	var err error
	dockerClient, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	httpClient = &http.Client{
		Timeout: httpClientTimeout * time.Second,
	}
}

func getFreePort(protocol string) int {
	switch protocol {
	case generic_utils.TCP:
		addr, err := net.ResolveTCPAddr(generic_utils.TCP, "localhost:0")
		if err != nil {
			panic(err)
		}

		return addr.Port
	case generic_utils.UDP:
		addr, err := net.ResolveUDPAddr(generic_utils.UDP, "localhost:0")
		if err != nil {
			panic(err)
		}

		return addr.Port
	default:
		panic(errors.Errorf("invalid port protocol: %s", protocol))
	}
}

func generatePortBindings(containerPorts nat.PortSet) (portMap nat.PortMap) {
	portMap = nat.PortMap{}

	for containerPort := range containerPorts {
		hostBinding := nat.PortBinding{
			HostIP:   defaultInterface,
			HostPort: strconv.Itoa(getFreePort(containerPort.Proto())),
		}
		portMap[containerPort] = []nat.PortBinding{hostBinding}
	}

	return
}

func startContainerHandler(w http.ResponseWriter, r *http.Request) {
	var containerInstance ContainerInstance
	http_utils.DecodeJSONRequestBody(r, &containerInstance)

	portBindings := generatePortBindings(containerInstance.Ports)

	containerConfig := container.Config{
		Image: containerInstance.ImageName,
	}

	hostConfig := container.HostConfig{
		PortBindings: portBindings,
	}

	cont, err := dockerClient.ContainerCreate(context.Background(), &containerConfig, &hostConfig,
		nil, "")
	if err != nil {
		panic(err)
	}

	err = dockerClient.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	log.Debugf("container %s started", cont.ID)

	serviceInstancePath := archimedes.GetServiceInstancePath(containerInstance.ServiceName, cont.ID)
	instanceDTO := archimedes.InstanceDTO{
		PortTranslation: portBindings,
	}

	req := http_utils.BuildRequest(http.MethodPost, localhost, serviceInstancePath, instanceDTO)
	statusCode, _ := http_utils.DoRequest(httpClient, req, nil)

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}
}
