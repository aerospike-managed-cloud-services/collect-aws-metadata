#!/usr/bin/env make

SHELL   := /bin/bash
PROG    := collect-aws-metadata
VERSION := $(shell tools/describe-version)
TARBALL := $(PROG)-$(VERSION)_$(GOOS)_$(GOARCH).tar.gz
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)

.PHONY: clean mock-service run-test test deps-test tarball


all: $(PROG)

$(PROG): collect.go go.mod go.sum
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build --ldflags="-X main.VERSION=$(VERSION)"

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
