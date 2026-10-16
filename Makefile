BIN := bin/zed-cheatsheet

.PHONY: run build list test vet clean

run:
	go run .

build:
	go build -o $(BIN) .

# Plain-text dump of the cheat sheet (no TUI needed).
list:
	go run . -list

test:
	go test ./...

vet:
	gofmt -l .
	go vet ./...

clean:
	rm -rf bin
