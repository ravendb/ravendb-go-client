
# go test -vet=off -covermode=atomic -coverprofile=coverage.txt

$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
$Env:LOG_ALL_REQUESTS = "true"
#$Env:ENABLE_FAILING_TESTS = "true"
#$Env:ENABLE_FLAKY_TESTS = "true"

go1.11beta3.exe test -vet=off -v -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestQuery$

#go1.11beta3.exe test -vet=off -v -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestAttachmentsSession$

