run:
	go run main.go -v

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
