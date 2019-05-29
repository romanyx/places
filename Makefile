SHELL := /bin/sh

test:
	go test -v -race `go list ./...`

linter:
	gometalinter `go list ./...`

start:
	docker-compose -f docker/docker-compose.yaml up
