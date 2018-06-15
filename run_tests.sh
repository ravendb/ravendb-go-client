#!/bin/bash

#go run -race cmd/test/*.go
#RAVENDB_JAVA_TEST_SERVER_PATH="./RavenDB/Server/Raven.Server"

# make Go http client use proxy
export HTTP_PROXY=http://localhost:8888

# TODO: for now not running with -race because fails with
# too many goroutines

#go test -race

go test
