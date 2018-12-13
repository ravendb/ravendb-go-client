
# go test -vet=off -covermode=atomic -coverprofile=coverage.txt

$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
$Env:LOG_ALL_REQUESTS = "true"
#$Env:ENABLE_FAILING_TESTS = "true"
#$Env:ENABLE_FLAKY_TESTS = "true"

go.exe test -timeout 20m -v ./tests
