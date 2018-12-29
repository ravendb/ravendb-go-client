#!/usr/bin/env pwsh

# go test -covermode=atomic -coverprofile=coverage.txt

$Env:VERBOSE_LOG = "true"
$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
# logs output of raven server to stdout, helpful for failing tests
#export LOG_RAVEN_SERVER=true
$Env:LOG_ALL_REQUESTS = "true"
$Env:ENABLE_FAILING_TESTS = "false"
$Env:ENABLE_FLAKY_TESTS = "false"

$wd = Join-Path "$PSScriptRoot" ".."
$ravdir = Join-Path $wd "RavenDB" "Server"

if ($IsMacOS) {
    $Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$ravdir/Raven.Server"
    $Env:RAVENDB_JAVA_TEST_CERTIFICATE_PATH="${wd}/certs/server.pfx"
    $Env:RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH="${wd}/certs/cert.pem"
    $Env:RAVENDB_JAVA_TEST_HTTPS_SERVER_URL="https://a.javatest11.development.run:8085"

} else {
    $Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$ravdir\Raven.Server.exe"
}

#$Env:RAVEN_GO_NO_DB_TESTS = "no"

go clean -testcache

#go.exe test -v -timeout 30s "-coverpkg=github.com/ravendb/ravendb-go-client" -covermode=atomic "-coverprofile=coverage.txt"  ./tests -run ^TestCachingOfDocumentInclude$

go test -v -race -timeout 50s ./tests -run ^TestAggregation$
