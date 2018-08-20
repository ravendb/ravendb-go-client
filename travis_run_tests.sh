#!/bin/bash
set -u -e -o pipefail -o xtrace

#export RAVENDB_JAVA_TEST_SERVER_PATH=./RavenDB/Server/Raven.Server

# comment/uncomment to turn additional logging on/off
export VERBOSE_LOG=true
export LOG_FAILED_HTTP_REQUESTS=true
export LOG_ALL_REQUESTS=true
export LOG_FAILED_HTTP_REQUESTS_DELAYED=true
#export ENABLE_FAILING_TESTS=true
#export ENABLE_FLAKY_TESTS=true

echo "TRAVIS_BUILD_DIR: $TRAVIS_BUILD_DIR"
echo "pwd:              `pwd`"
echo "GOPATH:           $GOPATH"


# this works in run_tests.sh
#go test -v -race  -vet=off -coverpkg github.com/ravendb/ravendb-go-client -covermode=atomic -coverprofile=coverage.txt ./tests

go test -v -race  -vet=off -coverpkg=all -covermode=atomic -coverprofile=coverage.txt ./tests
