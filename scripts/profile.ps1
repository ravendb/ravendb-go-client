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

# for running tests in a cluster, set to NODES_IN_CLUSTER to 3
# and KILL_SERVER_CHANCE to e.g. 10 (10%) and "SHUFFLE_CLUSTER_NODES" to true
$Env:NODES_IN_CLUSTER = "0"
$Env:KILL_SERVER_CHANCE = "0"
$Env:SHUFFLE_CLUSTER_NODES = "false"

$Env:ENABLE_PROFILING = "true"

rm ./tests/cpu.prof
go clean -testcache
go test -tags for_tests -v -timeout 80s ./tests -run ^TestAggressiveCaching$
ls -l ./tests/cpu.prof
