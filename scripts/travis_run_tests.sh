#!/bin/bash
set -u -e -o pipefail -o xtrace

# comment/uncomment to turn additional logging on/off
export VERBOSE_LOG=true
export LOG_FAILED_HTTP_REQUESTS=true
export LOG_ALL_REQUESTS=true
export LOG_FAILED_HTTP_REQUESTS_DELAYED=true
#export ENABLE_FAILING_TESTS=true
#export ENABLE_FLAKY_TESTS=true
# must use full absolute path because working directory is direrectory of
# ravendb server executable
wd=`pwd`
export RAVENDB_TEST_CERTIFICATE_PATH="${wd}/certs/server.pfx"
export RAVENDB_TEST_CA_PATH="${wd}/certs/ca.crt"
export RAVENDB_TEST_CLIENT_CERTIFICATE_PATH="${wd}/certs/cert.pem"
export RAVENDB_TEST_HTTPS_SERVER_URL="https://a.javatest11.development.run:8085"
export RAVENDB_SERVER_VERSION="5.1.11" # see .travis.yml

#export NODES_IN_CLUSTER="3"
#export KILL_SERVER_CHANCE="10"
#export SHUFFLE_CLUSTER_NODES=true

echo "pwd:              ${wd}"
echo "GOPATH:           ${GOPATH}"

# compile examples to catch mistakes there
go test -v -c ./examples

go test -tags for_tests -v -race -parallel 1 -timeout 30m -coverpkg=github.com/ravendb/ravendb-go-client -covermode=atomic -coverprofile=coverage.txt . ./tests
