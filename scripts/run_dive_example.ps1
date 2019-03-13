#!/usr/bin/env pwsh

Set-Location .\dive-into-raven
go run main.go log.go $args
Set-Location ..
