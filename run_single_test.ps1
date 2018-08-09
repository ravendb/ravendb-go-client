
# go test -vet=off -covermode=atomic -coverprofile=coverage.txt

$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"

go1.11beta3.exe test -vet=off -v -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestQuery$

#go1.11beta3.exe test -vet=off -v -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestAttachmentsSession$

