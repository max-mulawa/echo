.PHONY: build
build: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/echo ./cmd/echo/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/prime ./cmd/prime/main.go
test:
	go test