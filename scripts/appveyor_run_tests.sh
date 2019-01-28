#!/bin/bash
set -u -o pipefail -o xtrace

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
export RAVENDB_JAVA_TEST_CERTIFICATE_PATH="${wd}/certs/server.pfx"
export RAVENDB_JAVA_TEST_CA_PATH="${wd}/certs/ca.crt"
export RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH="${wd}/certs/cert.pem"
export RAVENDB_JAVA_TEST_HTTPS_SERVER_URL="https://a.javatest11.development.run:8085"
export RAVENDB_SERVER_VERSION="4.1.3" # see appveyor.yml

echo "pwd:              ${wd}"
echo "GOPATH:           ${GOPATH}"

go test -v -race -parallel 1 -timeout 20m -coverpkg=github.com/ravendb/ravendb-go-client -covermode=atomic -coverprofile=coverage.txt . ./tests
exitCode=$?

# upload coverage.txt to codecov.io
bash <(curl -s https://codecov.io/bash)

zip -r logs.zip logs
appveyor PushArtifact logs.zip

exit ${exitCode}
