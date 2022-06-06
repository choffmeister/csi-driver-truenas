.PHONY: *

test:
	go test $(shell go list ./... | grep -v e2e) -v

test-e2e:
	go test $(shell go list ./... | grep e2e) -v -timeout=60m

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

build:
	goreleaser release --rm-dist --skip-publish --snapshot

release:
	goreleaser release --rm-dist
