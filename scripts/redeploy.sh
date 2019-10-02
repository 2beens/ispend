#!/bin/bash
curl localhost:5001/harakiri
sleep 1s
go run cmd/main.go -port=5001 -env=p -logfile=ispend.log &