# Imports

Since you cant build project module-by-module with relative imports - they are absolute. But because of that you need explicitly(for now) change vendor name in imports. Line i use for this purpose is:

`find . -type f -name "*.go" -exec sed -i'' -e 's#github.com/ravendb#foo/ravendb#g' {} +`

# Running tests
test http package:

`go test -c -i github.com/ravendb-go-client/http`

`./http.go`

