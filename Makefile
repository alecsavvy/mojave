# Makefile for Sonata
.PHONY: dev setup deps tidy gen clean

# Run with hot reloading
# Usage: air -- run --home ./tmp/test-init
dev:
	./sonata

setup:
	chmod +x ./sonata

deps:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/air-verse/air@latest

tidy:
	go mod tidy

# Generate code from protobuf files
gen: clean
	buf generate
	make tidy

# Clean generated code
clean:
	rm -rf gen