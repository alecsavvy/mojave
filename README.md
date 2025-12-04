# Sonata

[![Go Reference](https://pkg.go.dev/badge/github.com/sonata-labs/sonata.svg)](https://pkg.go.dev/github.com/sonata-labs/sonata)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonata-labs/sonata?style=flat&v=1)](https://goreportcard.com/report/github.com/sonata-labs/sonata?style=flat&v=1)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The protocol for open audio distribution.

## Development

### Prerequisites

Install dependencies:

```bash
make deps
```

This installs [buf](https://buf.build/) for protobuf generation and [air](https://github.com/air-verse/air) for hot reloading.

### Running

Run the node:

```bash
./sonata run --home ./tmp/test-init
```

### Hot Reloading

For development with automatic rebuilds on file changes:

```bash
air -- run --home ./tmp/test-init
```

Air watches for changes to `.go`, `.toml`, and `.proto` files and automatically rebuilds and restarts the node.

### Initializing a Node

```bash
./sonata init --home ./tmp/test-init
```

### Generating Protobuf

```bash
make gen
```