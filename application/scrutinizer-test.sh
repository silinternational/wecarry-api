#!/usr/bin/env bash

# output everything
set -e
# exit on first error
set -x

# Install deps
go get ./...
go get -v -u github.com/gobuffalo/buffalo
go get -v -u github.com/gobuffalo/pop
go get -v -u github.com/gobuffalo/buffalo-pop
go get -v -u github.com/gobuffalo/suite
go get -v -u github.com/gobuffalo/httptest
go get -v -u github.com/markbates/grift

# install Buffalo cli
wget https://github.com/gobuffalo/buffalo/releases/download/v0.16.27/buffalo_0.16.27_Linux_x86_64.tar.gz
tar -xvzf buffalo_0.16.27_Linux_x86_64.tar.gz
chmod a+x buffalo

export LOG_LEVEL=fatal

# migrate db and run tests
buffalo-pop pop migrate up
./buffalo test -coverprofile=cover.out ./...
