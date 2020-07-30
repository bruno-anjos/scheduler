module github.com/bruno-anjos/scheduler

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/bruno-anjos/archimedes v0.0.0-20200730121306-72009db0eaa7
	github.com/bruno-anjos/solution-utils v0.0.0-20200729140846-0b732b78eb19
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
)

replace (
	github.com/bruno-anjos/archimedes v0.0.0-20200730121306-72009db0eaa7 => ./../archimedes
	github.com/bruno-anjos/solution-utils v0.0.0-20200729140846-0b732b78eb19 => ./../solution-utils
)
