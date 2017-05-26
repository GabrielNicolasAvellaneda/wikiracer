#!/bin/bash

set -e

docker >/dev/null 2>&1 || {
    echo "Docker must be installed"
    exit 1
}

docker images | grep -q racer || {
	./script/build.sh
}

docker ps | grep -q racer && {
	echo "Looks like wiki racer is already running..."
	exit 1
}

docker run -d -p 8081:8081 racer && {
	echo "|-------------------------------------------------------------|"
	echo "| Succesfully started wikiracer in docker :)                  |"
	echo "| GET/POST http://127.0.0.1:8081/api/v1/job to get/start jobs |"
	echo "|-------------------------------------------------------------|"
	exit 0
}

echo "something went wrong..."
exit 1
