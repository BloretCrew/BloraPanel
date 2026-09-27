# Invoke from the repository root: make -f cmd/log-helper/Makefile
GOOS ?= linux
GOARCH ?= amd64
OUTPUT ?= dist/blora-log-helper-$(GOOS)-$(GOARCH)
.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildvcs=false -trimpath -o $(OUTPUT) ./cmd/log-helper
