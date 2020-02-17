#!/usr/bin/env bash

if [ gqlgen/schema.graphql -nt gqlgen/generated.go ] || [ gqlgen/gqlgen.yml -nt gqlgen/generated.go ]
then
  echo -e "\e[31mERROR: gqlgen generated files are out of date\e[0m"
  exit 1
fi

whenavail testdb 5432 10 buffalo-pop pop migrate up
buffalo test
