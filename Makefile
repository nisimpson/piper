# Makefile for bee

pkg?=piper

.PHONY: test

test:
	go test -cover ./...

.PHONY: test-cover

test-cover:
	go test -coverprofile=/tmp/$(pkg).coverage.out ./...
	go tool cover -html=/tmp/$(pkg).coverage.out

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	golangci-lint run --disable-all --enable staticcheck --config staticcheck.yml
	govulncheck ./...
	SKIP_SHELL_TESTS=1 go test -race -buildvcs -vet=off ./...

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: no-dirty
no-dirty:
	git diff --exit-code

# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

.PHONY: lint
lint: tidy audit no-dirty

## push: push changes to the remote Git repository
.PHONY: push
push: lint
	git push

.PHONY: publish-throttle
publish-throttle:
	git tag -f throttle/v0.1.0
	git push origin throttle/v0.1.0
	@GOPROXY=proxy.golang.org go list -m github.com/nisimpson/piper/throttle@v0.1.0
