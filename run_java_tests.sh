#!/bin/bash

export RAVENDB_JAVA_TEST_SERVER_PATH=${HOME}/Documents/RavenDB/Server/Raven.Server

go run cmd/run_tests/run_tests.go -java

