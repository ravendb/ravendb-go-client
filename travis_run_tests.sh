#!/bin/bash

go test -race -covermode=atomic -coverprofile=coverage.txt
tar xvjf RavenDB.tar.bz2
rm -rf RavenDB.tar.bz2

