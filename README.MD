# How to install and run

Before this code is marged into github.com/ravendb/ravendb-go-client, the way to
install it on your computer and run the code:

```
mkdir -p $GOPATH/src/github.com/ravendb
cd $GOPATH/src/github.com/ravendb
rm -rf ravendb-go-client
git clone github.com/kjk/ravendb-go-client
cd ravendb-go-client
```

Basically in order to avoid renaming imports between github.com/ravendb/ravendb-go-client
and github.com/kjk/ravendb-go-client, my repo acts as if it is github.com/ravendb/ravendb-go-client
which means you can use `go get github.com/kjk/ravendb-go-client` but
have to place it in `$GOPATH/src/github.com/ravendb/ravendb-go-client` manually.

# Running tests

To run current tests do: `./run_tests.sh`.

It compiles and runs code in `cmd/test`.

Unsecured server should be running on http://localhost:9999.

So far this only has been tested on mac.
