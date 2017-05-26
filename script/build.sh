#!/bin/bash

set -e

docker >/dev/null 2>&1 || {
	echo "Docker must be installed"
	exit 1
}

docker build . -t racer
