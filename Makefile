.PHONY: all build run

# Default target: build the binary
all: build

# Build the binary
build:
	go build -o compose-viz ./cmd/compose-viz

# Run the binary with args and generate the image
run: build
	./compose-viz -v -f=docker-compose-example.yml && dot -Tpng composeGraph.dot  > output.png

open: run
	open output.png
