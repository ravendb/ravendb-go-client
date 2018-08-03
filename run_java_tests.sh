#!/bin/bash

currdir=`pwd`
export RAVENDB_JAVA_TEST_SERVER_PATH="${currdir}/RavenDB/Server/Raven.Server"
export HTTP_PROXY="yes"

go run cmd/run_java_tests/*.go
