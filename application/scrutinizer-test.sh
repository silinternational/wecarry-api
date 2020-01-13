#!/usr/bin/env bash

set -x
set -e

# Install deps
go get ./...
go get -v -u github.com/gobuffalo/pop
go get -v -u github.com/gobuffalo/packr/v2/packr2
go get -v -u github.com/gobuffalo/buffalo-pop
go get -v -u github.com/gobuffalo/suite
go get -v -u github.com/gobuffalo/httptest
go get -v -u github.com/markbates/grift

# migrate db and run tests
buffalo-pop pop migrate up
buffalo test -coverprofile=coverage.out ./...
