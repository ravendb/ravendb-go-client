#!/bin/bash

currdir=`pwd`
export RAVENDB_JAVA_TEST_SERVER_PATH="${currdir}/RavenDB/Server/Raven.Server"

go run cmd/run_tests/*.go -java
