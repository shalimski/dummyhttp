build:
	- GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/dummyhttp

test:
	- go test -v ./... -race

.PHONY: build test