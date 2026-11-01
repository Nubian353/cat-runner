.PHONY: run build test vet

run:
	go run .

build:
	go build -o cat-runner .

test:
	go test ./...

vet:
	go vet ./...
