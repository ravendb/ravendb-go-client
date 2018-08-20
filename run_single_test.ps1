
# go test -vet=off -covermode=atomic -coverprofile=coverage.txt

$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
$Env:LOG_ALL_REQUESTS = "true"
$Env:ENABLE_FAILING_TESTS = "false"
$Env:ENABLE_FLAKY_TESTS = "false"
$Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$PSScriptRoot\RavenDB\Server\Raven.Server.exe"

#go1.11beta3.exe test -v -vet=off -timeout 30s ./tests -run ^TestRavenDB8761$
go1.11beta3.exe test -v -vet=off -timeout 30s ./tests -run ^TestRavenDB903$

#go1.11beta3.exe test -vet=off -v -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestAttachmentsSession$

