## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':'

## init: initialize the project
.PHONY: install
init:
	@echo 'Init project...'
	go mod tidy
	cd ui && npm install

## build/web: build the cmd/web application
.PHONY: build/web
build/web:
	@echo 'Builing cmd/web...'
	go build -ldflags=-s -o=./bin/web ./cmd/web
	@echo 'Builing web assets...'
	cd ui && npm run build

## run/web: run the cmd/web application
.PHONY: run/web
run/web:
	go run ./cmd/web & cd ui && npm run dev
