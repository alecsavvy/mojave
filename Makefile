# Makefile for Sonata
.PHONY: dev setup deps tidy gen clean

dev:
	./sonata

setup:
	chmod +x ./sonata

deps:
	go install github.com/bufbuild/buf/cmd/buf@latest

tidy:
	go mod tidy

# Generate code from protobuf files
gen: clean
	buf generate
	make tidy

# Clean generated code
clean:
	rm -rf gen