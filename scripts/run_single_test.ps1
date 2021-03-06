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

$enableCluster = $false # change to $true to enable cluster setup

if ($enableCluster) {
    Write-Host "Cluster enabled"
    # for running tests in a cluster, set to NODES_IN_CLUSTER to 3
    # and KILL_SERVER_CHANCE to e.g. 10 (10%) and "SHUFFLE_CLUSTER_NODES" to true
    $Env:NODES_IN_CLUSTER = "3"
    $Env:KILL_SERVER_CHANCE = "0"
    $Env:SHUFFLE_CLUSTER_NODES = "true"
    $Env:LOG_TOPOLOGY = "true"
} else {
    Write-Host "Cluster not enabled"
    $Env:NODES_IN_CLUSTER = "0"
    $Env:KILL_SERVER_CHANCE = "0"
    $Env:SHUFFLE_CLUSTER_NODES = "false"
    $Env:LOG_TOPOLOGY = "false"
}

go clean -testcache

#go test -tags for_tests -v -timeout 30s "-coverpkg=github.com/ravendb/ravendb-go-client" -covermode=atomic "-coverprofile=coverage.txt"  ./tests -run ^TestCachingOfDocumentInclude$

go test -tags for_tests -v -race -timeout 60s ./tests -run ^TestNonNilTimeError$

#go test -tags for_tests -v -race -timeout 60s ./tests -run ^TestAggressiveCaching$

if (0) {
    # subscription worker tests
    go test -tags for_tests -v -race -timeout 30s ./tests -run ^TestSubscriptionsBasic$
    go test -tags for_tests -v -race -timeout 30s ./tests -run ^TestSecuredSubscriptionsBasic$
    go test -tags for_tests -v -race -timeout 30s ./tests -run ^TestRevisionsSubscriptions$
}

#go test -tags for_tests -v -race -timeout 60s ./tests -run ^TestAttachmentsRevisions$
#go test -tags for_tests -v -race -timeout 60s ./tests -run ^TestRevisions$
#go test -tags for_tests -v -race -timeout 60s ./tests -run ^TestAttachmentsSession$

