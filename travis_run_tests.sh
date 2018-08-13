#!/bin/bash
set -u -e -o pipefail -o xtrace

#export RAVENDB_JAVA_TEST_SERVER_PATH=./RavenDB/Server/Raven.Server

# comment/uncomment to turn additional logging on/off
export VERBOSE_LOG=true
export LOG_FAILED_HTTP_REQUESTS=true
export LOG_ALL_REQUESTS=true

go test -race -covermode=atomic -coverprofile=coverage.txt
