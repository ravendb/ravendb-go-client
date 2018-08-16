#!/bin/bash
set -u -e -o pipefail

# Disable database tests
export RAVEN_GO_NO_DB_TESTS=yes
go test -race
