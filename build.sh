#!/bin/bash

# sanity check: make sure that code and tests do compile

cd tests
go test -c
cd ..

go test -c
