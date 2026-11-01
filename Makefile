.PHONY: run build test vet check

run:
	go run .

build:
	go build -o cat-runner .

test:
	go test ./...

vet:
	go vet ./...

# то же, что проверяет CI
check:
	@unformatted="$$(gofmt -l .)"; if [ -n "$$unformatted" ]; then echo "gofmt:"; echo "$$unformatted"; exit 1; fi
	go vet ./...
	go test -race -count=1 ./...
