#!/bin/bash
curl localhost:5001/harakiri
sleep 2s
go run cmd/main.go -port=5001 -logfile=ispend.log &