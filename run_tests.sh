#!/bin/bash

#go run -race cmd/test/*.go
export RAVENDB_JAVA_TEST_SERVER_PATH=${HOME}/Documents/RavenDB/Server/Raven.Server
export HTTP_PROXY=http://localhost:8888
go test -race
