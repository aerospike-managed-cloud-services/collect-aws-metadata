#!/usr/bin/env make

SHELL 	       := /bin/bash
VERSION        := $(shell tools/describe-version)
PROG 	       := collect-aws-metadata
PROG_X86_64	   := $(PROG)-$(VERSION)-x86_64
PROG_ARM64	   := $(PROG)-$(VERSION)-arm64
TARBALL_X86_64 := $(PROG_X86_64).tar.gz
TARBALL_ARM64  := $(PROG_ARM64).tar.gz

.PHONY: clean mock-service run-test test deps-test tarball

all: $(PROG_X86_64)

$(PROG_X86_64): collect.go go.mod go.sum
	go build --ldflags="-X main.VERSION=$(VERSION)" -o $@
	mv $(PROG_X86_644) $(PROG) # Rename for testing

$(PROG_ARM64): collect.go go.mod go.sum
	GOARCH=arm64 GOOS=linux go build --ldflags="-X main.VERSION=$(VERSION)" -o $@
	mv $(PROG_ARM64) $(PROG) # Rename for testing

$(TARBALL_X86_64): $(PROG_X86_64)
	tar cfz $@ $^ && tar tvfz $@

$(TARBALL_ARM64): $(PROG_ARM64)
	tar cfz $@ $^ && tar tvfz $@

tarball-x86: $(TARBALL_X86_64) 

tarball-arm64: $(TARBALL_X86_64) 

clean:
	rm -f $(PROG_X86_64) $(TARBALL_X86_64) $(PROG_ARM64) $(TARBALL_ARM64) $(PROG)

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
