[![Linux build Status](https://travis-ci.org/kjk/ravendb-go-client.svg?branch=v4.0)](https://travis-ci.org/kjk/ravendb-go-client) [![Windows build status](https://ci.appveyor.com/api/projects/status/g2iida5te4gpgo7m?svg=true)](https://ci.appveyor.com/project/kjk/ravendb-go-client) [![codecov](https://codecov.io/gh/kjk/ravendb-go-client/branch/v4.0/graph/badge.svg)](https://codecov.io/gh/kjk/ravendb-go-client)

# How to install and run

Before this code is marged into github.com/ravendb/ravendb-go-client, the way to
install it on your computer and run the code:

```
mkdir -p $GOPATH/src/github.com/ravendb
cd $GOPATH/src/github.com/ravendb
rm -rf ravendb-go-client
git clone github.com/kjk/ravendb-go-client
cd ravendb-go-client
go get -v -u .\...
```

In order to avoid renaming imports between `github.com/ravendb/ravendb-go-client`
and `github.com/kjk/ravendb-go-client`, my repo acts as if it is `github.com/ravendb/ravendb-go-client`.

It means you can't use `go get github.com/kjk/ravendb-go-client`.

You have to manually place it in `$GOPATH/src/github.com/ravendb/ravendb-go-client`.

# Running tests

To run current tests do: `./run_tests.sh`.

I'm developing and testing mostly on Mac. The CI tests run on Linux.

# Running on Windows

Install as per Unix instructions.

When running `go get -v -u .\...` to get dependencies, ignore

```
# github.com/google/gopacket/pcap
exec: "gcc": executable file not found in %PATH%
```

Packet capturing is optional, to help debugging.

First, you have to install latest beta version of go, currently 1.11beta3 from https://golang.org/dl/#go1.11beta1
Otherwise you'll hit bug https://github.com/golang/go/issues/18468 (it should be backported to 1.10.4 as per https://github.com/golang/go/issues/25292).

When running `go test`, for now you must do `go test -vet=off`. This is to avoid another bug https://github.com/golang/go/issues/26390.

To run tests: `go test -vet=off`.

On Windows, if RavenDB server is not present locally, we'll download it to `RavenDB` directory.

Windows version doesn't capture traffic between HTTP client and server because `gopacket` package requires CGO support and pcap to work so I stubbed it out to be a no-op.
