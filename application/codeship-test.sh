#!/usr/bin/env bash

# output everything
set -e
# exit on first error
set -x

curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $GOPATH/bin latest

# https://github.com/securego/gosec#available-rules
# G104 ignore errors not checked
# G601 ignore implicit memory aliasing in for loop
gosec -exclude=G104,G601 -quiet ./...

whenavail testdb 5432 10 buffalo-pop pop migrate up
buffalo test
