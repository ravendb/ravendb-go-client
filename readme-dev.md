[![Linux build Status](https://travis-ci.org/ravendb/ravendb-go-client.svg?branch=master)](https://travis-ci.org/ravendb/ravendb-go-client) [![Windows build status](https://ci.appveyor.com/api/projects/status/rf326yoxl1uf444h/branch/master?svg=true)](https://ci.appveyor.com/project/ravendb/ravendb-go-client/branch/master)

# How to install and run

This is information for working on the library itself. For docs on how to use the library, see [readme.md](readme.md).

You need go 1.11 or later. Earlier versions have bugs that affect us (https://github.com/golang/go/issues/18468, https://github.com/golang/go/issues/26390) and don't support modules.

```
git clone https://github.com/ravendb/ravendb-go-client.git
cd ravendb-go-client
```

# Developing on Windows

To run all tests: `.\scripts\run_tests.ps1`.

On Windows, if RavenDB server is not present locally, we'll download it to `RavenDB` directory.

# Developing on Mac

To avoid writing helper scripts twice, many are written in PowerShell. You can install PowerShell on mac using [Homebrew](https://brew.sh/): `brew cask install powershell` (see https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-macos?view=powershell-6 for up-to-date information).

For running HTTPS tests you must import `certs/ca.crt` as trusted certificate:

* double-click on `certs/ca.crt` file. That opens `Keychain Access` system app.
* lick on `Certificates` category, double-click on `a.javatest11.development.run` certificate.
* this opens a dialog box. In `Trust` section select `Always Trust` drop-down item.

To run all tests: `./scripts/run_tests.ps`

More dev information:
* [porting_notes.md](porting_notes.md)
* [handling_maps.md](handling_maps.md)
