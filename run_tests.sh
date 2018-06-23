#!/bin/bash

#go run -race cmd/test/*.go
#RAVENDB_JAVA_TEST_SERVER_PATH="./RavenDB/Server/Raven.Server"

# make Go http client use proxy
export HTTP_PROXY=http://localhost:8888
#export HTTP_PROXY=

# TODO: for now not running with -race because fails with:
# "race: limit on 8192 simultaneously alive goroutines is exceeded, dying"
# in requestExecutorTest_failsWhenServerIsOffline when also running
# debugging proxy.
# For now locally run with proxy but without -race.
# On CI we run -race but not proxy

#go test -race

go test

#go test -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestRequestExecutor$