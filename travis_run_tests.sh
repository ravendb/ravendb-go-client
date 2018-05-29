#!/bin/bash

# Disable database tests
export RAVEN_GO_NO_DB_TESTS=yes
go test -race -covermode=atomic -coverprofile=coverage.txt

