run:
	CGO_ENABLED=0 go run main.go server -P -v 2 --port-forward

# Will build binaries using GoReleaser.
# Install GoReleaser on forehand. https://goreleaser.com/install/
# For more information about the commands, read the docs https://goreleaser.com/customization/
build:
	goreleaser build --clean --snapshot

clean:
	rm -f kubedock
	rm -rf dist
	go mod tidy
	rm -f coverage.out
	go clean -testcache

cloc:
	cloc --exclude-dir=vendor,node_modules,dist,_notes,_archive .

fmt:
	find ./internal -type f -name \*.go -exec gofmt -s -w {} \;
	go fmt ./...

test:
	CGO_ENABLED=0 go vet ./...
	CGO_ENABLED=0 go test ./... -cover

lint:
	golint ./internal/...
	# errcheck ./internal/... ./cmd/...

cover:
	CGO_ENABLED=0 go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

deps:
	go install golang.org/x/lint/golint@latest
	go install github.com/kisielk/errcheck@latest

.PHONY: run build clean cloc fmt test lint cover deps
