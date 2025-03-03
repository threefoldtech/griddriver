#!/bin/bash
set -ex

go get -u ./...
go mod tidy
