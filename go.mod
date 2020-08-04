module github.com/bruno-anjos/scheduler

go 1.13

require (
	github.com/bruno-anjos/archimedes v0.0.0-20200730160527-37e36e2f1583
	github.com/bruno-anjos/solution-utils v0.0.0-20200803160423-4cf841cde3d3
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
)

replace (
	github.com/bruno-anjos/archimedes v0.0.0-20200730160527-37e36e2f1583 => ./../archimedes
	github.com/bruno-anjos/solution-utils v0.0.0-20200803160423-4cf841cde3d3 => ./../solution-utils
)
