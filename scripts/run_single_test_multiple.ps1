#!/usr/bin/env pwsh

# on mac install powershell: https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-macos?view=powershell-6

# this runs a single test 10 times (or until first failure).
# helps find flaky tests

$Env:VERBOSE_LOG = "true"
$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
# logs output of raven server to stdout, helpful for failing tests
#export LOG_RAVEN_SERVER=true
$Env:LOG_ALL_REQUESTS = "true"
$Env:ENABLE_FAILING_TESTS = "false"
$Env:ENABLE_FLAKY_TESTS = "false"

if ($IsMacOS) {
    $ravdir = "RavenDB/Server"
    $Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$ravdir/Raven.Server"
} else {
    $ravdir = Join-Path -Path "$PSScriptRoot" -ChildPath ".." -Resolve
    $ravdir = "$ravdir\RavenDB\Server"
    $Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$ravdir\Raven.Server.exe"
}

For ($i=0; $i -lt 10; $i++) {

    go clean -testcache
    go test -parallel 1 -v -timeout 50s ./tests -run ^TestBulkInserts$

    if ($lastexitcode -ne 0) {
        exit
    }
}
