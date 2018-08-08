
# go test -vet=off -covermode=atomic -coverprofile=coverage.txt

# force running tests even if code didn't change
go1.11beta3 clean -testcache

go1.11beta3 test -vet=off -v -race -timeout 30s github.com/ravendb/ravendb-go-client -run ^TestQuery$
