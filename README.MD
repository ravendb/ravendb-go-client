[![Build Status](https://travis-ci.org/kjk/ravendb-go-client.svg?branch=v4.0)](https://travis-ci.org/kjk/ravendb-go-client)

[![codecov](https://codecov.io/gh/kjk/ravendb-go-client/branch/v4.0/graph/badge.svg)](https://codecov.io/gh/kjk/ravendb-go-client)

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

In order to avoid renaming imports between `github.com/ravendb/ravendb-go-client`
and `github.com/kjk/ravendb-go-client`, my repo acts as if it is `github.com/ravendb/ravendb-go-client`.

It means you can't use `go get github.com/kjk/ravendb-go-client`.

You have to manually place it in `$GOPATH/src/github.com/ravendb/ravendb-go-client`.

# Running tests

To run current tests do: `./run_tests.sh`.

I'm developing (and testing) on Mac. The CI tests are run on Linux.

Not tested on Windows yet.
