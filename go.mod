module github.com/bruno-anjos/scheduler

go 1.13

require (
	github.com/bruno-anjos/archimedes v0.0.0-20200730160527-37e36e2f1583
	github.com/bruno-anjos/solution-utils v0.0.0-20200803160206-562c9f14e46c
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
)

replace (
	github.com/bruno-anjos/archimedes v0.0.0-20200730160527-37e36e2f1583 => ./../archimedes
	github.com/bruno-anjos/solution-utils v0.0.0-20200731153528-f4f5b5285b7d => ./../solution-utils
)
