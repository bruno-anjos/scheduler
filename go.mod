module github.com/bruno-anjos/scheduler

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/bruno-anjos/archimedes v0.0.0-20200804153633-d07ca32d62f3
	github.com/bruno-anjos/archimedesHTTPClient v0.0.0-20200804154915-4a52ba818e68 // indirect
	github.com/bruno-anjos/solution-utils v0.0.0-20200804140242-989a419bda22
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
)

replace (
	github.com/bruno-anjos/archimedes v0.0.0-20200804153633-d07ca32d62f3 => ./../archimedes
	github.com/bruno-anjos/solution-utils v0.0.0-20200804140242-989a419bda22 => ./../solution-utils
)
