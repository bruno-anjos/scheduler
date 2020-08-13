package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	archimedes "github.com/bruno-anjos/archimedes/api"
	deployer "github.com/bruno-anjos/deployer/api"
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
	typeInstanceToContainerMapKey   = string
	typeInstanceToContainerMapValue = string
)

const (
	httpClientTimeout    = 10
	stopContainerTimeout = 10

	networkName = "scheduler-network"
)

var (
	dockerClient        *client.Client
	networkId           string
	httpClient          *http.Client
	instanceToContainer sync.Map

	stopContainerTimeoutVar = stopContainerTimeout * time.Second
)

func init() {
	log.SetLevel(log.DebugLevel)

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

	networkConfig := types.NetworkCreate{
		CheckDuplicate: false,
		Attachable:     false,
	}

	networks, err := dockerClient.NetworkList(context.Background(), types.NetworkListOptions{})

	exists := false
	for _, network := range networks {
		if network.Name == networkName {
			networkId = network.ID
			exists = true
			break
		}
	}

	if !exists {
		var resp types.NetworkCreateResponse
		resp, err = dockerClient.NetworkCreate(context.Background(), networkName, networkConfig)
		if err != nil {
			panic(err)
		}

		networkId = resp.ID
		log.Debug("created network with id ", networkId)
	} else {
		log.Debug("network ", networkName, " already exists")
	}

	log.SetLevel(log.InfoLevel)
}

func getFreePort(protocol string) string {
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

		natPort, err := nat.NewPort(utils.TCP, strconv.Itoa(l.Addr().(*net.TCPAddr).Port))
		if err != nil {
			panic(err)
		}

		return natPort.Port()
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

		natPort, err := nat.NewPort(utils.UDP, strconv.Itoa(l.LocalAddr().(*net.UDPAddr).Port))
		if err != nil {
			panic(err)
		}

		return natPort.Port()
	default:
		panic(errors.Errorf("invalid port protocol: %s", protocol))
	}
}

func generatePortBindings(containerPorts nat.PortSet) (portMap nat.PortMap) {
	portMap = nat.PortMap{}

	for containerPort := range containerPorts {

		hostBinding := nat.PortBinding{
			HostIP:   utils.LocalhostAddr,
			HostPort: getFreePort(containerPort.Proto()),
		}
		portMap[containerPort] = []nat.PortBinding{hostBinding}
	}

	return
}

func startInstanceHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("handling start instance")

	var containerInstance api.ContainerInstanceDTO
	err := json.NewDecoder(r.Body).Decode(&containerInstance)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if containerInstance.ServiceName == "" || containerInstance.ImageName == "" {
		log.Errorf("invalid container instance: %v", containerInstance)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go startContainerAsync(&containerInstance)
}

func startContainerAsync(containerInstance *api.ContainerInstanceDTO) {
	portBindings := generatePortBindings(containerInstance.Ports)

	//
	// Create container and get containers id in response
	//
	instanceId := containerInstance.ServiceName + "-" + utils.RandomString(10)

	log.Debugf("instance %s has following portBindings: %+v", instanceId, portBindings)

	serviceIdEnvVar := utils.ServiceEnvVarName + "=" + containerInstance.ServiceName
	instanceIdEnvVar := utils.InstanceEnvVarName + "=" + instanceId

	envVars := []string{serviceIdEnvVar, instanceIdEnvVar}
	envVars = append(envVars, containerInstance.EnvVars...)

	out, err := dockerClient.ImagePull(context.Background(), containerInstance.ImageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer func() {
		err = out.Close()
		if err != nil {
			panic(err)
		}
	}()

	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		panic(err)
	}

	containerConfig := container.Config{
		Env:   envVars,
		Image: containerInstance.ImageName,
	}

	hostConfig := container.HostConfig{
		NetworkMode:  "bridge",
		PortBindings: portBindings,
	}

	cont, err := dockerClient.ContainerCreate(context.Background(), &containerConfig, &hostConfig,
		nil, instanceId)
	if err != nil {
		log.Error(dockerClient.ClientVersion())
		panic(err)
	}

	err = dockerClient.NetworkConnect(context.Background(), networkId, cont.ID, nil)
	if err != nil {
		panic(err)
	}

	//
	// Add container instance to archimedes
	//
	serviceInstancePath := archimedes.GetServiceInstancePath(containerInstance.ServiceName, instanceId)
	instanceDTO := archimedes.InstanceDTO{
		PortTranslation: portBindings,
		Static:          containerInstance.Static,
		Local:           true,
	}

	req := http_utils.BuildRequest(http.MethodPost, archimedes.DefaultHostPort, serviceInstancePath, instanceDTO)
	statusCode, _ := http_utils.DoRequest(httpClient, req, nil)

	if statusCode != http.StatusOK {
		err = dockerClient.ContainerStop(context.Background(), cont.ID, &stopContainerTimeoutVar)
		if err != nil {
			log.Error(err)
		}
		log.Fatalf("got status code %d while adding instances to archimedes", statusCode)
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

	deployerPath := deployer.GetRegisterDeploymentInstancePath(containerInstance.ServiceName, cont.ID)
	req = http_utils.BuildRequest(http.MethodPost, deployer.DefaultHostPort, deployerPath, nil)
	status, _ := http_utils.DoRequest(httpClient, req, nil)
	if status != http.StatusOK {
		log.Fatalf("got error code %d while registering instance", status)
	}

	log.Debugf("container %s started for instance %s", cont.ID, instanceId)
}

func stopInstanceHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("handling delete instance")

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
	go stopContainerAsync(instanceId, contId)
}

func stopContainerAsync(instanceId, contId string) {
	err := dockerClient.ContainerStop(context.Background(), contId, &stopContainerTimeoutVar)
	if err != nil {
		panic(err)
	}

	log.Debugf("deleted instance %s corresponding to container %s", instanceId, contId)
}

func stopAllInstancesHandler(_ http.ResponseWriter, _ *http.Request) {
	deleteAllInstances()
}

func deleteAllInstances() {
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
