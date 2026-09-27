export GOCACHE ?= /tmp/blora-go-build
export GOMODCACHE ?= /tmp/blora-go-mod
export GOPATH ?= /tmp/blora-go
export GOFLAGS ?= -buildvcs=false
export npm_config_cache ?= /tmp/blora-npm-cache

.PHONY: build test check web sdk windows fixture package
build:
	go build -trimpath -o dist/blora-master ./cmd/master
	go build -trimpath -o dist/blora-daemon ./cmd/daemon
test:
	go test -race ./...
check:
	go vet ./...
web:
	cd web && npm ci && npm run build
sdk:
	cd sdk && npm ci && npm run check && npm run build
	cd sdk/examples/reference-app && npm ci && npm run build && npm run package && npm run package:fixtures
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o dist/blora-master.exe ./cmd/master
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o dist/blora-daemon.exe ./cmd/daemon
fixture: build
	go build -trimpath -o dist/blora-devfixture ./cmd/devfixture
	exec ./dist/blora-devfixture
package: build windows
	go build -trimpath -o dist/blora-extension-sign ./cmd/extension-sign
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o dist/blora-extension-sign.exe ./cmd/extension-sign
	cd web && npm run build
	cd sdk && npm run build
	cd sdk/examples/reference-app && npm run package && npm run package:fixtures
	python3 scripts/third-party-notices.py
	python3 scripts/package.py --version "$${BLORA_VERSION:-development}"
