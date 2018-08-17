#!/bin/bash

$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
$Env:LOG_ALL_REQUESTS = "true"
#$Env:ENABLE_FAILING_TESTS = "true"
#$Env:ENABLE_FLAKY_TESTS = "true"
$Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$PSScriptRoot/RavenDB/Server/Raven.Server"
$Env:HTTP_PROXY = "http://localhost:8888"

go run .\cmd\run_java_tests\run_java_tests.go
