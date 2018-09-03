#!/bin/bash

# Helper for running a test I'm currently working on
# Faster than running all tests

function check() {
    if [ ! -f ./capturer ]; then
        echo "./capturer not found!"
        echo "Run ./build_capturer.sh to create it"
        exit 1
    fi

    if [ ! -f ./pcap_convert ]; then
        echo "./pcap_convert not found!"
        echo "Run ./build_capturer.sh to create it"
        exit 1
    fi
}

set -o xtrace

export VERBOSE_LOG=true
export LOG_HTTP_REQUEST_SUMMARY=true
export LOG_FAILED_HTTP_REQUESTS=true
# logs output of raven server to stdout, helpful for failing tests
# export LOG_RAVEN_SERVER=true
#export PCAP_CAPTURE=true
export LOG_ALL_REQUESTS=true
#export ENABLE_FAILING_TESTS=true
#export ENABLE_FLAKY_TESTS=true

# cd tests

# force running tests even if code didn't change
go clean -testcache

# go test -race -vet=off -v -timeout 60s github.com/ravendb/ravendb-go-client/tests -run ^TestCrud$

#go test -race -vet=off -v -timeout 60s github.com/ravendb/ravendb-go-client/tests -run ^TestRavenDB5669$ ./tests

go test -v -race -vet=off -run ^TestCustomSerialization$ ./tests
