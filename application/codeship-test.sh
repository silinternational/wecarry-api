#!/usr/bin/env bash

whenavail testdb 5432 10 buffalo-pop pop migrate up
buffalo test
