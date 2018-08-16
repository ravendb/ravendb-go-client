#!/bin/bash

#RAVENDB_JAVA_TEST_SERVER_PATH="./RavenDB/Server/Raven.Server"

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

# comment/uncomment to turn additional logging on/off
export VERBOSE_LOG=true
export LOG_FAILED_HTTP_REQUESTS=true
#export PCAP_CAPTURE=true
export LOG_ALL_REQUESTS=true
export LOG_FAILED_HTTP_REQUESTS_DELAYED=true
#export ENABLE_FAILING_TESTS=true
#export ENABLE_FLAKY_TESTS=true

cd tests

# go test -race
go test -race -covermode=atomic -coverprofile=coverage.txt
