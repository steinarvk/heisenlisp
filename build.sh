#!/bin/bash
pigeon grammar/heisenlisp.peg > gen/parser/parser.go && go test && go build
