#!/bin/bash
set -u -e -o pipefail

#RAVENDB_JAVA_TEST_SERVER_PATH=./RavenDB/Server/Raven.Server
go test -race -covermode=atomic -coverprofile=coverage.txt

