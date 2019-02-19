#!/usr/bin/env pwsh
# on mac install powershell: https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-macos?view=powershell-6

$Env:VERBOSE_LOG = "true"
$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
# logs output of raven server to stdout, helpful for failing tests
#export LOG_RAVEN_SERVER=true
$Env:LOG_ALL_REQUESTS = "true"
$Env:ENABLE_FAILING_TESTS = "false"
$Env:ENABLE_FLAKY_TESTS = "false"
$Env:ENABLE_NORTHWIND_TESTS = "true"

# $IsMacOS is only defined in powershell 6, but it happens to work
# in windows with powershell 5 because it's not defined at all, so false
if ($IsMacOS) {
    $wd = Join-Path -Path "$PSScriptRoot" -ChildPath ".." -Resolve
    $ravdir = "${wd}/RavenDB/Server"
    $Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$ravdir/Raven.Server"
    $Env:RAVENDB_JAVA_TEST_CERTIFICATE_PATH = "${wd}/certs/server.pfx"
    $env:RAVENDB_JAVA_TEST_CA_PATH = "${wd}/certs/ca.crt"
    $Env:RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH = "${wd}/certs/cert.pem"
    $Env:RAVENDB_JAVA_TEST_HTTPS_SERVER_URL = "https://a.javatest11.development.run:8085"
}
else {
    $ravdir = Join-Path -Path "$PSScriptRoot" -ChildPath ".." -Resolve
    $ravdir = "$ravdir\RavenDB\Server"
    $Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$ravdir\Raven.Server.exe"
}

go clean -testcache

#go test -tags for_tests -v -timeout 30s "-coverpkg=github.com/ravendb/ravendb-go-client" -covermode=atomic "-coverprofile=coverage.txt"  ./tests -run ^TestCachingOfDocumentInclude$

go test -tags for_tests -v -race -timeout 60s ./tests -run ^TestGo1$
