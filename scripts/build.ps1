# sanity check: make sure that code and tests do compile
# compiling tests will also compile the code

go test -v -c .
go test -v -c .\tests
