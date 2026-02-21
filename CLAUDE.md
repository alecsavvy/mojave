# Mojave

Decentralized music distribution protocol. Go + CometBFT + OpenTDF + BitTorrent + Casbin + Goja.

## Project status

Early stage — scaffold with placeholder implementations. Architecture is designed and documented in `docs/`. Implementation is next.

## Architecture (read these first)

- `PROTOCOL.md` — **client-facing interface contract** (API, crypto, content access, payment). Copy this into client repos for their LLMs.
- `docs/README.md` — why this exists, how to read the docs, open design questions
- `docs/architecture.md` — full system design: four planes (consensus, storage, encryption, policy), all transaction types, all flows
- `docs/storage.md` — two PebbleDB stores: chain store (ABCI state) and local store (validator-specific)
- `docs/content.md` — on-disk file layout, BitTorrent integration, reconciliation loop
- `docs/economics.md` — MOJ token, fees, validator rewards, bootstrapping
- `docs/governance.md` — validator/oracle elections, takedowns, jurisdictional compliance
- `docs/diagrams/` — mermaid diagrams referenced from the docs

## Code layout

```
cmd/mojave/       CLI entrypoint (cobra)
commands/         cobra commands (root, start)
app/              ABCI application (placeholder)
store/            chain state store (placeholder for PebbleDB)
content/          content-addressed file store (placeholder)
server/           external API server (placeholder)
config/           configuration loading
proto/            protobuf service definitions
docs/             architecture and design documentation
```

## Tech stack

- **Language**: Go (1.25+)
- **Consensus**: CometBFT (ABCI state machine, BFT consensus, p2p reactors)
- **State store**: PebbleDB (two instances — chain store via CometBFT, local store for validator data)
- **Encryption**: OpenTDF (DEK/KEK wrapping, `.flac.tdf` containers)
- **File replication**: BitTorrent (CID-based rendezvous for encrypted blobs and images)
- **Storage abstraction**: gocloud.dev (blob.Bucket — local disk, S3, GCS, Azure)
- **Policy engine**: Casbin (RBAC/ABAC/ACL with custom on-chain adapter)
- **Programmable runtime**: Goja (sandboxed ECMAScript 5.1+ in Go, for policies and proofs)
- **API**: ConnectRPC (proto-generated, HTTP/1.1+JSON) + GraphQL (gqlgen, autobind to proto types)
- **CLI**: cobra
- **Proto**: protobuf for transactions, p2p messages, wire format
- **DDEX types**: imported from separate `ddex-proto` Go module
- **Desktop client**: Tauri v2 (Rust backend, React frontend via Bun + Vite)
- **Crypto**: Ed25519 (signing/identity), X25519 (device-scoped encryption, ECDH DEK wrapping)
- **Token**: MOJ (1 MOJ = 10^9 grains)

## Key conventions

- All audio normalized to FLAC, all images normalized to PNG
- Content files: `audio/{cid[0:2]}/{cid[2:4]}/{cid}.flac.tdf`, `images/{cid[0:2]}/{cid[2:4]}/{cid}.png`
- Chain store keys prefixed by domain: `uploads/`, `releases/`, `entitlements/`, `policies/`, `deks/`, `grants/`, `accounts/`
- Local store keys prefixed: `local/deks/`, `local/processing/`, `local/sync/`
- All on-chain values in protobuf. All balances/fees in grains (uint64).
- Ed25519 for signing, X25519 for encryption. Never use Ed25519 for encryption directly.
- Wrapped DEKs are always ciphertext. Raw DEKs only exist during upload processing.
- Validators are elected (staking + social vote), not just staked.

## Commands

```bash
go build -o mojave ./cmd/mojave    # build
go run ./cmd/mojave start          # run (placeholder)
go test ./...                      # test
```

## Important constraints

- The upload validator temporarily holds unencrypted audio — trust assumption, not a bug
- Goja scripts must be deterministic (no Date.now, no Math.random, no async, no eval)
- Policy evaluation must produce identical results across all validators for consensus
- Validator X25519 keys are registered on-chain; consumer X25519 keys are device-scoped and never on-chain
- The uploader's wrapped DEK is always on-chain; validator wrapped DEKs are off-chain (local store, p2p distributed)
