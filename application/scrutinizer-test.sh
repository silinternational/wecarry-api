#!/usr/bin/env bash

set -x
set -e

# Install deps
go get ./...
go get -v -u github.com/gobuffalo/buffalo
go get -v -u github.com/gobuffalo/pop
go get -v -u github.com/gobuffalo/packr/v2/packr2
go get -v -u github.com/gobuffalo/buffalo-pop
go get -v -u github.com/gobuffalo/suite
go get -v -u github.com/gobuffalo/httptest
go get -v -u github.com/markbates/grift

# install Buffalo cli
wget https://github.com/gobuffalo/buffalo/releases/download/v0.15.3/buffalo_0.15.3_Linux_x86_64.tar.gz
tar -xvzf buffalo_0.15.3_Linux_x86_64.tar.gz
chmod a+x buffalo

# migrate db and run tests
buffalo-pop pop migrate up
./buffalo test -coverprofile=cover.out ./...
