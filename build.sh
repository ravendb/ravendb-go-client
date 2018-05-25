#!/bin/bash

go build -race cmd/test/*.go
go test -c
