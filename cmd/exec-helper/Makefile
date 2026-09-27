# Invoke from the repository root: make -f cmd/exec-helper/Makefile
GOARCH ?= amd64
OUTPUT ?= dist/blora-exec-helper-linux-$(GOARCH)
.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -buildvcs=false -trimpath -o $(OUTPUT) ./cmd/exec-helper
