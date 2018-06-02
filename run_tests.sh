#!/bin/bash

#go run -race cmd/test/*.go
export RAVENDB_JAVA_TEST_SERVER_PATH=${HOME}/Documents/RavenDB/Server/Raven.Server

# make Go http client use proxy
# https://golang.org/pkg/net/http/#ProxyFromEnvironment
#export HTTP_PROXY=http://localhost:8888
export NO_PROXY=true
go test -race
