run:
	go run main.go -v 2

build:
	CGO_ENABLED=0 go build -ldflags \
		"-X github.com/joyrex2001/kubedock/internal/config.Date=`date -u +%Y%m%d-%H%M%S`  \
		 -X github.com/joyrex2001/kubedock/internal/config.Build=`git rev-list -1 HEAD`   \
		 -X github.com/joyrex2001/kubedock/internal/config.Version=`git describe --tags`  \
		 -X github.com/joyrex2001/kubedock/internal/config.Image=joyrex2001/kubedock:`git describe --tags | cut -d- -f1`" \
		 -o kubedock

docker:
	docker build . -t joyrex2001/kubedock:latest

clean:
	rm -f kubedock
	go mod tidy

cloc:
	cloc --exclude-dir=vendor,node_modules,dist,_notes .

fmt:
	find ./internal -type f -name \*.go -exec gofmt -s -w {} \;
	go fmt ./...

test:
	go vet ./...
	go test ./... -cover

lint:
	golint ./internal/...
	errcheck ./internal/... ./cmd/...

cover:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out
	
deps:
	go get -u golang.org/x/lint/golint

.PHONY: run build docker clean cloc fmt test lint cover deps