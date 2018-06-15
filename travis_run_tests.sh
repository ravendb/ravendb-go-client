#!/bin/bash
set -u -e -o pipefail

#RAVENDB_JAVA_TEST_SERVER_PATH=./RavenDB/Server/Raven.Server

# TODO: for now disabling -race because fails with "too many goroutines"
#go test -race -covermode=atomic -coverprofile=coverage.txt
go test -covermode=atomic -coverprofile=coverage.txt
