[![Linux build Status](https://travis-ci.org/ravendb/ravendb-go-client.svg?branch=v4.0)](https://travis-ci.org/ravendb/ravendb-go-client) [![Windows build status](https://ci.appveyor.com/api/projects/status/rf326yoxl1uf444h/branch/v4.0?svg=true)](https://ci.appveyor.com/project/ravendb/ravendb-go-client/branch/v4.0)

# How to install and run

You need go 1.11 or later. You can download it from https://golang.org/dl/
Earlier versions have bugs that affect us (https://github.com/golang/go/issues/18468, https://github.com/golang/go/issues/26390) and dont' support modules.


```
git clone https://github.com/ravendb/ravendb-go-client.git
cd ravendb-go-client
```

# Running tests

To run current tests: `./scripts/run_tests.sh`.

I'm developing and testing mostly on Mac. The CI tests run on Linux.

# Running on Windows

Install as per Unix instructions.

To run current tests: `.\scripts\run_tests.ps1`.

On Windows, if RavenDB server is not present locally, we'll download it to `RavenDB` directory.

# Running on Mac

For running https tests must import `certs/ca.crt` as trusted certificate.

Double-click on `certs/ca.crt` file. That opens `Keychain Access` system app.

Click on `Certificates` category, double-click on `a.javatest11.development.run` certificate.

This opens a dialog box. In `Trust` section select `Always Trust` drop-down item.
