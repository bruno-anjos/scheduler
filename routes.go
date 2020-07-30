package main

import (
	"fmt"
	"net/http"

	"github.com/bruno-anjos/solution-utils/http_utils"
)

// Route names
const (
	startContainerName = "START_CONTAINER"
	stopContainerName  = "STOP_CONTAINER"
)

// Paths
const (
	PrefixPath = "/scheduler"

	ContainerPath = "/containers/%s"
)

// Path variables
const (
	containerIdPathVar = "containerId"
)

var (
	_containerIdPathVarFormatted = fmt.Sprintf(http_utils.PathVarFormat, containerIdPathVar)

	containerRoute = fmt.Sprintf(ContainerPath, _containerIdPathVarFormatted)
)

var routes = []http_utils.Route{
	{
		Name:        startContainerName,
		Method:      http.MethodPost,
		Pattern:     containerRoute,
		HandlerFunc: startContainerHandler,
	},

	/*{
		Name:        stopContainerName,
		Method:      http.MethodDelete,
		Pattern:     containerRoute,
		HandlerFunc: stopContainerHandler,
	},*/
}
