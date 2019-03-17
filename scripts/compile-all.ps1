#!/usr/bin/env pwsh

# sanity check: make sure that code and tests do compile
# compiling tests will also compile the code

go test -tags for_tests -v -c ./tests
go test -v -c ./examples
go test -v -c ./dive-into-raven

