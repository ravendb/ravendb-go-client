#!/bin/bash

# Helper for running a test I'm currently working on
# Faster than running all tests

# Uncomment to make Go http client use proxy
#export HTTP_PROXY=http://localhost:8888
export HTTP_PROXY=

# uncomment for more verbose logging
export VERBOSE_LOG=true

# antidote to test caching
go clean -testcache

# go test -race -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestAttachmentsSession$

go test -v -race -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestIndexesFromClient$

