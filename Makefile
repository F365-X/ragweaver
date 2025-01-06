BINARY_NAME := ./bin/ragweaver

# Dependent libraries (here, bmatcuk/doublestar is explicitly specified as an example)
DEPS := github.com/bmatcuk/doublestar/v4

# Default target
all: build

# Get and check dependencies
deps:
	go mod tidy

# Build the Go binary
build: deps
	go build -o $(BINARY_NAME) main.go

# Run the binary
run:
	# Usage: make run ARGS='path/to/repo -p path/to/preamble.txt'
	go run main.go $(ARGS)

# Run tests (if any)
test:
	go test ./...

# Remove the binary
clean:
	rm -f $(BINARY_NAME)

# Install the binary
install: build
	sudo install -m 0755 $(BINARY_NAME) /usr/local/bin/

.PHONY: all deps build run test clean install
