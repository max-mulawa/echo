.PHONY: build
build: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/echo ./cmd/echo/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/prime ./cmd/prime/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/means ./cmd/means/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/chat ./cmd/chat/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/kvstore ./cmd/kvstore/main.go
test:
	go test ./...