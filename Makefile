.PHONY: build build-local build-release build-all lint lint-go lint-goreleaser lint-codecov lint-docs test clean setup generate-mocks

build: build-local

build-local: lint generate-mocks
	mkdir -p bin
	go build -o bin/projectkit ./cmd/projectkit
	go build -o bin/projectkit-mcp ./cmd/projectkit-mcp

# Release build (all platforms via GoReleaser)
build-release: lint
	goreleaser build --snapshot --clean

# Lint
lint: lint-go lint-goreleaser lint-codecov

lint-go:
	golangci-lint run --fix

lint-goreleaser:
	goreleaser check

lint-codecov:
	curl --fail --silent --show-error --data-binary @codecov.yml https://codecov.io/validate >/dev/null

test: lint generate-mocks
	go test -tags=coverage -coverprofile=coverage.out ./...

# Generate mocks
generate-mocks:
	mockery

projectkit-setup: build
	./bin/projectkit setup --log-level=debug

# Clean
clean:
	rm -rf bin/ dist/
