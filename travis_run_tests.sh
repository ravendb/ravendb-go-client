#!/bin/bash

#RAVENDB_JAVA_TEST_SERVER_PATH=./RavenDB/Server/Raven.Server
go test -race -covermode=atomic -coverprofile=coverage.txt

