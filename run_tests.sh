#!/bin/bash

#RAVENDB_JAVA_TEST_SERVER_PATH="./RavenDB/Server/Raven.Server"

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

set -o xtrace

# uncomment for more verbose logging
export VERBOSE_LOG=true
export PCAP_CAPTURE=true
export LOG_FAILED_HTTP_REQUESTS=true

# go test -race
go test -covermode=atomic -coverprofile=coverage.txt
