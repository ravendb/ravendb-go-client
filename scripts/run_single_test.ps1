
# go test -vet=off -covermode=atomic -coverprofile=coverage.txt

$Env:VERBOSE_LOG = "true"
$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
# logs output of raven server to stdout, helpful for failing tests
#export LOG_RAVEN_SERVER=true
$Env:LOG_ALL_REQUESTS = "true"
$Env:ENABLE_FAILING_TESTS = "false"
$Env:ENABLE_FLAKY_TESTS = "false"

$Env:RAVENDB_JAVA_TEST_SERVER_PATH = "$PSScriptRoot\..\RavenDB\Server\Raven.Server.exe"

#$Env:RAVEN_GO_NO_DB_TESTS = "no"

go.exe clean -testcache

#go.exe test -v -timeout 30s "-coverpkg=github.com/ravendb/ravendb-go-client" -covermode=atomic "-coverprofile=coverage.txt"  ./tests -run ^TestCachingOfDocumentInclude$

#go.exe test -v -timeout 30s ./tests -run ^TestCachingOfDocumentInclude$
#go.exe test -v -timeout 30s ./tests -run ^TestWhatChanged$

go.exe test -parallel 1 -v -timeout 50s ./tests -run ^TestRavenDB8761$

#go.exe test -vet=off -v -timeout 15s github.com/ravendb/ravendb-go-client -run ^TestAttachmentsSession$
