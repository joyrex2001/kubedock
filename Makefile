run:
	go run main.go server -P -v 2

build:
	CGO_ENABLED=0 go build -ldflags \
		"-X github.com/joyrex2001/kubedock/internal/config.Date=`date -u +%Y%m%d-%H%M%S`  \
		 -X github.com/joyrex2001/kubedock/internal/config.Build=`git rev-list -1 HEAD`   \
		 -X github.com/joyrex2001/kubedock/internal/config.Version=`git describe --tags`  \
		 -X github.com/joyrex2001/kubedock/internal/config.Image=joyrex2001/kubedock:`git describe --tags | cut -d- -f1`" \
		 -o kubedock

gox:
	CGO_ENABLED=0 gox -os="linux darwin windows" -arch="amd64" \
		-output="dist/kubedock_`git describe --tags`_{{.OS}}_{{.Arch}}" -ldflags \
		"-X github.com/joyrex2001/kubedock/internal/config.Date=`date -u +%Y%m%d-%H%M%S`  \
		 -X github.com/joyrex2001/kubedock/internal/config.Build=`git rev-list -1 HEAD`   \
		 -X github.com/joyrex2001/kubedock/internal/config.Version=`git describe --tags`  \
		 -X github.com/joyrex2001/kubedock/internal/config.Image=joyrex2001/kubedock:`git describe --tags | cut -d- -f1`"

docker:
	docker build . -t joyrex2001/kubedock:latest

clean:
	rm -f kubedock
	rm -rf dist
	go mod tidy
	rm -f coverage.out

cloc:
	cloc --exclude-dir=vendor,node_modules,dist,_notes .

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
	go get -u golang.org/x/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u github.com/mitchellh/gox
	go get -u github.com/tcnksm/ghr

.PHONY: run build gox docker clean cloc fmt test lint cover deps