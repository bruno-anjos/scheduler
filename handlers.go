package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	archimedes "github.com/bruno-anjos/archimedes/api"
	"github.com/bruno-anjos/scheduler/api"
	utils "github.com/bruno-anjos/solution-utils"
	"github.com/bruno-anjos/solution-utils/http_utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type (
	typeInstanceToContainerMapKey = string
	typeInstanceToContainerMapValue = string
)

const (
	httpClientTimeout    = 10
	stopContainerTimeout = 10
)

var (
	dockerClient        *client.Client
	httpClient          *http.Client
	instanceToContainer sync.Map

	stopContainerTimeoutVar = stopContainerTimeout * time.Second
)

func init() {
	var err error
	dockerClient, err = client.NewEnvClient()
	if err != nil {
		log.Error("unable to create docker client")
		panic(err)
	}

	httpClient = &http.Client{
		Timeout: httpClientTimeout * time.Second,
	}

	instanceToContainer = sync.Map{}
}

func getFreePort(protocol string) int {
	switch protocol {
	case utils.TCP:
		addr, err := net.ResolveTCPAddr(utils.TCP, "0.0.0.0:0")
		if err != nil {
			panic(err)
		}

		l, err := net.ListenTCP(utils.TCP, addr)
		if err != nil {
			panic(err)
		}

		defer func() {
			err = l.Close()
			if err != nil {
				panic(err)
			}
		}()

		return l.Addr().(*net.TCPAddr).Port
	case utils.UDP:
		addr, err := net.ResolveUDPAddr(utils.UDP, "0.0.0.0:0")
		if err != nil {
			panic(err)
		}

		l, err := net.ListenUDP(utils.UDP, addr)
		if err != nil {
			panic(err)
		}

		defer func() {
			err = l.Close()
			if err != nil {
				panic(err)
			}
		}()

		return l.LocalAddr().(*net.UDPAddr).Port
	default:
		panic(errors.Errorf("invalid port protocol: %s", protocol))
	}
}

func generatePortBindings(containerPorts nat.PortSet) (portMap nat.PortMap) {
	portMap = nat.PortMap{}

	for containerPort := range containerPorts {

		hostBinding := nat.PortBinding{
			HostIP:   utils.DefaultInterface,
			HostPort: strconv.Itoa(getFreePort(containerPort.Proto())),
		}
		portMap[containerPort] = []nat.PortBinding{hostBinding}
	}

	return
}

func startInstanceHandler(w http.ResponseWriter, r *http.Request) {
	var containerInstance api.ContainerInstanceDTO
	err := json.NewDecoder(r.Body).Decode(&containerInstance)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if containerInstance.ServiceName == "" || containerInstance.ImageName == "" {
		log.Errorf("invalid container instance: %v", containerInstance)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	portBindings := generatePortBindings(containerInstance.Ports)

	//
	// Create container and get containers id in response
	//
	containerConfig := container.Config{
		Image: containerInstance.ImageName,
	}

	hostConfig := container.HostConfig{
		PortBindings: portBindings,
	}

	cont, err := dockerClient.ContainerCreate(context.Background(), &containerConfig, &hostConfig,
		nil, "")
	if err != nil {
		log.Error(dockerClient.ClientVersion())
		panic(err)
	}

	//
	// Add container instance to archimedes
	//
	instanceId := containerInstance.ServiceName + "-" + cont.ID[:10]
	serviceInstancePath := archimedes.GetServiceInstancePath(containerInstance.ServiceName, instanceId)
	instanceDTO := archimedes.InstanceDTO{
		PortTranslation: portBindings,
	}

	req := http_utils.BuildRequest(http.MethodPost, archimedes.DefaultHostPort, serviceInstancePath, instanceDTO)
	statusCode, _ := http_utils.DoRequest(httpClient, req, nil)

	if statusCode != http.StatusOK {
		err = dockerClient.ContainerStop(context.Background(), cont.ID, &stopContainerTimeoutVar)
		if err != nil {
			log.Error(err)
		}
		w.WriteHeader(statusCode)
		return
	}

	//
	// Spin container up
	//
	err = dockerClient.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	instanceToContainer.Store(instanceId, cont.ID)

	log.Debugf("container %s started for instance %s", cont.ID, instanceId)

	http_utils.SendJSONReplyOK(w, instanceId)
}

func stopInstanceHandler(w http.ResponseWriter, r *http.Request) {
	instanceId := http_utils.ExtractPathVar(r, instanceIdPathVar)

	if instanceId == "" {
		log.Errorf("no instance provided", instanceId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, ok := instanceToContainer.Load(instanceId)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	contId := value.(typeInstanceToContainerMapValue)
	err := dockerClient.ContainerStop(context.Background(), contId, &stopContainerTimeoutVar)
	if err != nil {
		panic(err)
	}
}

func stopAllInstancesHandler(_ http.ResponseWriter, _ *http.Request) {
	log.Debugf("stopping all containers")

	instanceToContainer.Range(func(key, value interface{}) bool {
		instanceId := value.(typeInstanceToContainerMapKey)
		contId := value.(typeInstanceToContainerMapValue)

		log.Debugf("stopping instance %s (container %s)", instanceId, contId)

		err := dockerClient.ContainerStop(context.Background(), contId, &stopContainerTimeoutVar)
		if err != nil {
			log.Warnf("error while stopping instance %s (container %s): %s", instanceId, contId, err)
			return true
		}

		return true
	})

	containers, err := dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, containerListed := range containers {
		log.Warnf("deleting orphan container %s", containerListed.ID)
		err = dockerClient.ContainerStop(context.Background(), containerListed.ID, &stopContainerTimeoutVar)
		if err != nil {
			log.Errorf("error stopping orphan container %s: %s", containerListed.ID, err)
		}
	}
}


