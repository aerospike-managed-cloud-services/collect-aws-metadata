#!/usr/bin/env make

SHELL := /usr/bin/bash

.PHONY: clean mock-service run-test


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
		--base-url http://localhost:8000 \
		--textfiles-path=/tmp/collect-aws
