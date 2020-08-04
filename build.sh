#!/bin/bash

set -e

env CGO_ENABLED=0 GOOS=linux go build -o scheduler .
docker build -t brunoanjos/scheduler:latest .
docker push brunoanjos/scheduler:latest