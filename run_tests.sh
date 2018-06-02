#!/bin/bash

#go run -race cmd/test/*.go
#RAVENDB_JAVA_TEST_SERVER_PATH="./RavenDB/Server/Raven.Server"

# make Go http client use proxy
export HTTP_PROXY=http://localhost:8888

go test -race
