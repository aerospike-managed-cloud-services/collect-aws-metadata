#!/usr/bin/env make

SHELL := /usr/bin/bash

.PHONY: clean mock-service run-test test deps-test


all: collect-aws-metadata


clean:
	rm -f collect-aws-metadata


mock-service:
	. $(VIRTUAL_ENV)/bin/activate && cd test \
		&& uvicorn mock_metadata:app \
			--header server:EC2ws \
			--reload


collect-aws-metadata: collect.go go.mod
	go build


run-test: collect.go go.mod
	go run . \
		--base-url=http://localhost:8000 \
		--metric-prefix=amcs_ \
		--textfiles-path=/tmp/collect-aws

deps-test:
	go install github.com/dave/courtney

test: deps-test
	# go test -v -coverprofile cover.out .
	courtney .
	go tool cover -func coverage.out
	go tool cover -html coverage.out -o coverage.html
	# intentionally fail test just to see what GHA does

test-100pct: deps-test
	courtney -e .
