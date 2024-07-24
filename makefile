#!/usr/bin/env make

SHELL 	:= /opt/homebrew/bin/bash
PROG 	:= collect-aws-metadata
VERSION := v1.1.0
TARBALL := $(PROG)-arm64-$(VERSION).tar.gz

.PHONY: clean mock-service run-test test deps-test tarball


all: $(PROG)

$(PROG): collect.go go.mod go.sum
	GOARCH=arm64 GOOS=linux go build --ldflags="-X main.VERSION=$(VERSION)"

$(TARBALL): $(PROG)
	tar cfz $@ $^ && tar tvfz $@

tarball: $(TARBALL)

clean:
	rm -f $(PROG) $(TARBALL)

mock-service:
	. $(VIRTUAL_ENV)/bin/activate && cd test \
		&& uvicorn mock_metadata:app \
			--header server:EC2ws \
			--reload

run-test: collect.go go.mod
	go run . \
		--base-url=http://localhost:8000 \
		--metric-prefix=amcs_ \
		--textfiles-path=/tmp/collect-aws

deps-test:
	go install github.com/dave/courtney

test: deps-test
	courtney .
	go tool cover -func coverage.out
	go tool cover -html coverage.out -o coverage.html

test-100pct: deps-test
	courtney -e .
