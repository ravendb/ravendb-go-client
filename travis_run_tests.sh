#!/bin/bash
set -u -e -o pipefail

#RAVENDB_JAVA_TEST_SERVER_PATH=./RavenDB/Server/Raven.Server

VERBOSE_LOG=true
#go test -race -covermode=atomic -coverprofile=coverage.txt
go test -race -covermode=atomic -coverprofile=coverage.txt
