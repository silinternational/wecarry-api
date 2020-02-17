#!/usr/bin/env bash

if [ gqlgen/schema.graphql -nt gqlgen/generated.go ] || [ gqlgen/gqlgen.yml -nt gqlgen/generated.go ]
then
  echo "gqlgen generated files are out of date"
  exit 1
fi

whenavail testdb 5432 10 buffalo-pop pop migrate up
buffalo test
