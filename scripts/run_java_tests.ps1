#!/bin/bash

$Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$PSScriptRoot\..\RavenDB\Server\Raven.Server.exe"
$Env:HTTP_PROXY = "http://localhost:8888"

go run .\cmd\run_java_tests\run_java_tests.go
