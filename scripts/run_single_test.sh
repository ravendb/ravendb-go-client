#!/bin/bash

# Helper for running a test I'm currently working on
# Faster than running all tests

set -o xtrace

export VERBOSE_LOG=true
export LOG_HTTP_REQUEST_SUMMARY=true
export LOG_FAILED_HTTP_REQUESTS=true
# logs output of raven server to stdout, helpful for failing tests
#export LOG_RAVEN_SERVER=true
export LOG_ALL_REQUESTS=true
#export ENABLE_FAILING_TESTS=true
#export ENABLE_FLAKY_TESTS=true

# must use full absolute path because working directory is direrectory of
# ravendb server executable
wd=`pwd`
export RAVENDB_JAVA_TEST_CERTIFICATE_PATH="${wd}/certs/server.pfx"
export RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH="${wd}/certs/cert.pem"
export RAVENDB_JAVA_TEST_HTTPS_SERVER_URL="https://a.javatest11.development.run:8085"

# cd tests

# force running tests even if code didn't change
go clean -testcache

# go test -race -vet=off -v -timeout 60s github.com/ravendb/ravendb-go-client/tests -run ^TestCrud$

#go test -race -vet=off -v -timeout 60s github.com/ravendb/ravendb-go-client/tests -run ^TestRavenDB5669$ ./tests

#go test -v -race -vet=off -run "^TestC.*$" ./tests
go test -v -race -vet=off -run ^TestLazy$ ./tests
