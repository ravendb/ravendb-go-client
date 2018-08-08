#!/bin/bash

# Helper for running a test I'm currently working on
# Faster than running all tests

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

# force running tests even if code didn't change
go clean -testcache

export PCAP_CAPTURE=true

# go test -race -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestAttachmentsSession$

go test -vet=off -v -race -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestQuery$
