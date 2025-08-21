.PHONY: build
build:
	go build -v ./cmd/service

.DEFAULT_GOAL := build
