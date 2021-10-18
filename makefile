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
