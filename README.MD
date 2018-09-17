[![Linux build Status](https://travis-ci.org/ravendb/ravendb-go-client.svg?branch=v4.0)](https://travis-ci.org/ravendb/ravendb-go-client) [![Windows build status](https://ci.appveyor.com/api/projects/status/rf326yoxl1uf444h/branch/v4.0?svg=true)](https://ci.appveyor.com/project/ravendb/ravendb-go-client/branch/v4.0)

# How to install and run

```
go get github.com/ravendb/ravendb-go-client
cd $GOPATH/src/ravendb-go-client
# get dependencies
go get -v -u .\...
```

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

# Running on Mac

For running https tests must import `certs/ca.crt` as trusted certificate.

Double-click on `certs/ca.crt` file. That opens `Keychain Access` system app.

Click on `Certificates` category, double-click on `a.javatest11.development.run` certificate.

This opens a dialog box. In `Trust` section select `Always Trust` drop-down item.
