#!/bin/bash

currdir=`pwd`
export RAVENDB_JAVA_TEST_SERVER_PATH="${currdir}/RavenDB/Server/Raven.Server"
export HTTP_PROXY=http://localhost:8888

go run cmd/run_java_tests/*.go
