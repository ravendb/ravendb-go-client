#!/bin/bash
set -u -e -o pipefail

rm -rf RavenDB.tar.bz2

wget -O RavenDB.tar.bz2 https://daily-builds.s3.amazonaws.com/RavenDB-4.0.3-osx-x64.tar.bz2

# TODO: daily seems broken
# wget -O RavenDB.tar.bz2 https://hibernatingrhinos.com/downloads/RavenDB%20for%20OSX/latest?buildType=nightly

rm -rf ./RavenDB
tar xvjf RavenDB.tar.bz2
