@rem default test -timeout is 10m which sometimes is not enough

go test -tags for_tests -v -parallel 1 -timeout 20m "-coverpkg=github.com/ravendb/ravendb-go-client" -covermode=atomic "-coverprofile=coverage.txt" . ./tests

set testerrorlevel=%errorlevel%

7z a logs.zip logs
appveyor PushArtifact logs.zip

exit /b %testerrorlevel%
