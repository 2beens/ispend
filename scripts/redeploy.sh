#!/bin/bash
port=5001
logFile=ispend.log

function rebirth() {
    local port="$1"
    local logFile="$2"

    echo "Killing the server [localhost:${port}]..."
    curl localhost:${port}/harakiri
    echo
    sleep 1s
    echo "Getting born again..."
    echo "run cmd/main.go -port=${port} -logfile=${logFile} &"
    go run cmd/main.go -port=${port} -logfile=${logFile} &
    echo "Server got reborn"
}

rebirth ${port} ${logFile}
