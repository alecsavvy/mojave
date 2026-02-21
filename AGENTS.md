# Mojave — Agent Instructions

Decentralized music distribution protocol built in Go. CometBFT for consensus, OpenTDF for encryption, BitTorrent for file replication, Casbin + Goja for access control.

## Documentation

Read `docs/README.md` first — it explains the full doc structure, reading order, and lists 17 open design questions. The architecture is designed but implementation is early.

**For client repos:** `PROTOCOL.md` is the external interface contract (API, crypto, content access, payment flows). Copy it into client repos so their LLMs have full protocol context without needing access to this repo.

| Doc | What it covers |
|-----|---------------|
| `docs/architecture.md` | Full system design — four planes, state machine, transaction types, all flows, API layer |
| `docs/storage.md` | Two PebbleDB stores — chain store key spaces, local store key spaces, rebuilding |
| `docs/content.md` | File layout (`.flac.tdf` + `.png`), directory sharding, BitTorrent, reconciliation |
| `docs/economics.md` | MOJ token (grains), fees, rewards, staking, genesis, bootstrapping |
| `docs/governance.md` | Validator/oracle elections, takedowns, jurisdictional compliance |
| `docs/diagrams/` | 16 mermaid diagrams referenced from the docs |

## Project structure

```
cmd/mojave/       CLI entrypoint
commands/         cobra commands
app/              CometBFT ABCI application (placeholder)
store/            chain state store — will be PebbleDB via CometBFT
content/          content-addressed file store — will use gocloud.dev
server/           API server — will expose ConnectRPC + GraphQL
config/           configuration
proto/            protobuf service definitions
docs/             architecture documentation
```

## Build and run

```bash
go build -o mojave ./cmd/mojave
go run ./cmd/mojave start
go test ./...
```

## Tech stack

Go 1.25+, CometBFT, PebbleDB, OpenTDF, BitTorrent, gocloud.dev, Casbin, Goja, ConnectRPC, gqlgen, cobra, protobuf. Desktop client: Tauri v2 (Rust + React). Crypto: Ed25519 (signing), X25519 (device-scoped encryption via ECDH).

## Coding conventions

- All audio normalized to FLAC before encryption, all images to PNG
- Content paths are deterministic from CID: `audio/{hex[0:2]}/{hex[2:4]}/{hex}.flac.tdf`
- Chain store key prefixes: `uploads/`, `releases/`, `entitlements/`, `policies/`, `deks/`, `grants/`, `accounts/`, `validators/`, `oracles/`, `takedowns/`, `flags/`
- Local store key prefixes: `local/deks/`, `local/processing/`, `local/sync/`
- All on-chain values protobuf-encoded. All monetary values in grains (uint64, 10^9 grains = 1 MOJ)
- Ed25519 for identity/signing, X25519 for encryption. Never encrypt with Ed25519 directly
- Wrapped DEKs are ciphertext. Raw DEKs only exist transiently during upload processing
- Goja scripts must be deterministic — no Date.now(), Math.random(), async, eval, or external I/O
- Validators elected via social vote + staking, not staking alone

## Key architectural decisions

- Two PebbleDB instances per validator: chain store (consensus, identical) and local store (instance-specific)
- Uploader's wrapped DEK always on-chain (recovery). Validator wrapped DEKs off-chain (local store, p2p)
- Consumer encryption keys are device-scoped X25519 — generated locally, never on-chain
- Replication set (who stores files) and DEK holder set (who can grant access) are independently configurable by the uploader
- BitTorrent for bulk file transfer, CometBFT reactors for small p2p messages (wrapped DEKs)
- Casbin for structured IAM, Goja for programmable policies and proof generation
- GraphQL is the read layer, protobuf + CometBFT is the write layer (single `sendTx` mutation)
- Oracles (elected, separate from validators) handle copyright takedowns via DEK removal
