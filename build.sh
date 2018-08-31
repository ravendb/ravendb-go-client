#!/bin/bash

# sanity check: make sure that code and tests do compile
# compiling tests will also compile the code

cd tests
go test -v -c
