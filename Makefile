GO ?= $(shell command -v go 2> /dev/null)
MACHINE = $(shell uname -m)

GOFLAGS ?= $(GOFLAGS:)

export GO111MODULE=on

## Checks the code style, tests, builds and bundles.
all: check-style dist 

## Runs govet and gofmt against all packages.
.PHONY: check-style
check-style: govet lint
	@echo Checking for style guide compliance

## Runs lint against all packages.
.PHONY: lint
lint:
	@echo Running lint
	env GO111MODULE=off $(GO) get -u golang.org/x/lint/golint
	golint -set_exit_status ./...
	@echo lint success

## Runs govet against all packages.
.PHONY: vet
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

## Builds and thats all :)
.PHONY: dist
dist:	build

.PHONY: build
build: ## Build 
	@echo Building 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build \
	     -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) \
	     -a -installsuffix cgo -o build/influx-sim .
#GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GO) build -ldflags \
#     -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) \
#     -a -installsuffix cgo -o build/influx-sim .

