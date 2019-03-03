#!/usr/bin/env pwsh

Set-Location .\examples
go run main.go log.go $args
Set-Location ..
