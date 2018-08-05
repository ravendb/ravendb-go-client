#!/bin/bash
set -o xtrace

# Helper for running a test I'm currently working on
# Faster than running all tests

# uncomment for more verbose logging
export VERBOSE_LOG=true

# force running tests even if code didn't change
go clean -testcache

export PCAP_CAPTURE=true

# go test -race -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestAttachmentsSession$

go test -v -race -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestIndexesFromClient$
