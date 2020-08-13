module github.com/bruno-anjos/scheduler

go 1.13

require (
	github.com/bruno-anjos/archimedes v0.0.2
	github.com/bruno-anjos/deployer v0.0.1
	github.com/bruno-anjos/solution-utils v0.0.1
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
)

replace (
	github.com/bruno-anjos/archimedes v0.0.2 => ./../archimedes
	github.com/bruno-anjos/deployer v0.0.1 => ./../deployer
	github.com/bruno-anjos/solution-utils v0.0.1 => ./../solution-utils
)
