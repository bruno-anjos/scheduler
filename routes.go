package main

import (
	"fmt"
	"net/http"

	"github.com/bruno-anjos/scheduler/api"
	"github.com/bruno-anjos/solution-utils/http_utils"
)

// Route names
const (
	startInstanceName = "START_INSTANCE"
	stopInstanceName  = "STOP_INSTANCE"
	stopAllInstancesName  = "STOP_ALL_INSTANCES"
)

const (
	instanceIdPathVar = "instanceId"
)

var (
	_instanceIdPathVarFormatted = fmt.Sprintf(http_utils.PathVarFormat, instanceIdPathVar)

	instancesRoute = api.InstancesPath
	instanceRoute  = fmt.Sprintf(api.InstancePath, _instanceIdPathVarFormatted)
)

var routes = []http_utils.Route{
	{
		Name:        startInstanceName,
		Method:      http.MethodPost,
		Pattern:     instancesRoute,
		HandlerFunc: startInstanceHandler,
	},

	{
		Name:        stopInstanceName,
		Method:      http.MethodDelete,
		Pattern:     instanceRoute,
		HandlerFunc: stopInstanceHandler,
	},

	{
		Name:        stopAllInstancesName,
		Method:      http.MethodDelete,
		Pattern:     instancesRoute,
		HandlerFunc: stopAllInstancesHandler,
	},
}
