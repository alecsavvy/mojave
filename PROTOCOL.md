# Mojave Protocol — Client Reference

This document is the external interface contract for building clients against the Mojave network. It covers everything a client developer needs: authentication, API, content access, crypto, and payment. It does not cover validator internals, consensus, or state machine implementation — those are in `docs/`.

Copy this file into your client repo's root so your LLM has full protocol context without needing access to the Mojave repo.

---

## What Mojave is

A decentralized music distribution protocol. Validators (elected, accountable operators) store encrypted audio, enforce access policies, and serve content. Clients download encrypted files, request decryption keys, and play locally. The model is iTunes (own your library), not Spotify (stream from a service).

## Authentication

Clients authenticate with **Ed25519 keypairs**. No sessions, no cookies, no passwords.

- The user's identity is their Ed25519 public key.
- All requests that require identity are signed with the Ed25519 private key.
- Wallet adapters (Phantom Connect, WalletConnect, etc.) handle key management. Social login (Google, Apple) via Phantom provides Ed25519 keys without the user managing raw keys.
- To authenticate to the API: sign a challenge message with the Ed25519 key, send the signature. The validator verifies and scopes responses to that key's entitlements.

## Device encryption keys

Each device generates its own **X25519 keypair** locally. This is separate from the Ed25519 identity key.

- Store the X25519 private key in the device keychain / secure storage.
- The X25519 public key is included in access requests (see below).
- The validator wraps the decryption key (DEK) against this device's X25519 public key.
- A wrapped DEK from one device is useless on another — different X25519 key.
- Multi-device is trivial: each device requests its own wrapped DEK independently.

**Key derivation summary:**

| Key | Algorithm | Where it lives | What it does |
|-----|-----------|---------------|-------------|
| Identity key | Ed25519 | Wallet (self-custody or custodial) | Sign transactions, access requests, authenticate |
| Device encryption key | X25519 | Device keychain | Unwrap DEKs via ECDH. Never leaves the device. |

## API

Validators expose two APIs. Use whichever fits your client.

### GraphQL (recommended for UIs)

Flexible query language. Ask for exactly what you need. GraphiQL endpoint available for interactive exploration.

**Read queries** — rich and typed:

| Query | What it returns |
|-------|----------------|
| What do I own? | List of CIDs where your key has an `owner` entitlement |
| What can I play? | List of CIDs where your key has been granted access |
| Get release metadata | DDEX ERN, FLAC CID, encrypted CID, image CIDs, status |
| Get cover art | PNG image (served directly, no access gate) |
| Get access history | History of access grants — when, what, which validator |
| Get policies for my content | Access policies and Casbin state for your catalog |

**Write mutation** — single mutation for all writes:

```graphql
mutation {
  sendTx(signedTransaction: "<base64-encoded signed protobuf tx>") {
    hash
    code
    log
  }
}
```

All writes are signed protobuf transactions submitted through CometBFT. The GraphQL layer is read-only except for this single `sendTx` mutation.

### ConnectRPC (recommended for programmatic access)

gRPC-compatible, works over HTTP/1.1 + JSON. No proxy needed for browsers. Generated from the same `.proto` files as the validators. Good for SDKs, CLI tools, and typed programmatic access.

### Content serving

Validators serve content files directly over HTTP:

```
GET /content/audio/{encrypted_cid}  →  encrypted .flac.tdf blob
GET /content/images/{image_cid}     →  PNG image
```

Audio files are ciphertext — no access check needed to download. The access check happens when requesting a DEK. Images are unencrypted and public.

## Content files

| Type | Extension | Encrypted | Format |
|------|-----------|-----------|--------|
| Audio | `.flac.tdf` | Yes (OpenTDF) | FLAC wrapped in a TDF container |
| Image | `.png` | No | PNG |

All audio is normalized to FLAC. All images are normalized to PNG. No format negotiation needed.

## Core flows

### 1. Browse library

```
Device                          Validator API
  |                                |
  |-- sign challenge (Ed25519) --->|
  |<-- authenticated session ------|
  |                                |
  |-- GraphQL: "what can I play?" -|
  |   (signed by Ed25519)          |
  |                                |
  |<-- list of CIDs + DDEX metadata
  |    + image CIDs (cover art)    |
  |                                |
  |-- HTTP GET cover art PNGs ---->|
  |<-- PNG images -----------------|
```

### 2. Download content

Two paths depending on client type:

| Client | Transport for `.flac.tdf` | Transport for `.png` | Local storage |
|--------|--------------------------|---------------------|---------------|
| Desktop (Tauri) | BitTorrent via `librqbit` | HTTP GET | User's filesystem |
| Browser | HTTP GET from validator | HTTP GET | OPFS (Origin Private File System) |

The encrypted CID (on-chain, in the upload record) is the BitTorrent rendezvous identifier. Desktop clients join the swarm using this CID and download from validators, good samaritans, or other consumers who are seeding.

`.flac.tdf` files are ciphertext — safe to download and store without a DEK. They sit on disk as opaque blobs until decrypted.

### 3. Request access (get a DEK)

```
Device                          DEK Holder Validator
  |                                |
  |-- access request:              |
  |   { cid,                       |
  |     device_x25519_pubkey }     |
  |   signed by Ed25519 --------->|
  |                                |
  |   validator checks policy      |
  |   validator wraps DEK to       |
  |   device's X25519 pubkey       |
  |                                |
  |<-- wrapped DEK ----------------|
```

The access request includes:
- `cid` — which content you want
- `device_x25519_pubkey` — this device's X25519 public key
- Ed25519 signature over the request

If the validator redirects you (it's in the replication set but not the DEK holder set), follow the redirect to a DEK holder validator and repeat.

### 4. Decrypt and play

```
1. Unwrap the DEK:
   - ECDH: device X25519 private key × ephemeral public key (from wrapped DEK)
   - Derive symmetric wrapping key
   - Decrypt the wrapped DEK → raw DEK

2. Decrypt the .flac.tdf:
   - Open the TDF container
   - Decrypt the payload with the DEK
   - Result: raw FLAC audio

3. Play the FLAC
```

All crypto happens on-device. The validator never sees the raw DEK or unencrypted audio during consumption.

### 5. Purchase content

Content purchases are embedded in the access request. The access policy (set by the content owner) determines the price.

The `GrantAccess` transaction atomically:
- Transfers MOJ (in grains) from the consumer's account to the content owner's account
- Records the access grant on-chain (audit trail)
- Returns the wrapped DEK to the device

If the content is free (`public` policy), no payment is needed.

### 6. Offline playback

Once a device has both:
- The `.flac.tdf` file (downloaded via BitTorrent or HTTP)
- A valid (non-expired) wrapped DEK

It can play offline indefinitely. No network needed. Permanently purchased content has non-expiring DEKs — true offline ownership. Subscription content has expiring DEKs that require periodic refresh.

## Protobuf transactions

All writes are protobuf messages signed by the submitter and submitted via `sendTx`. The client:

1. Constructs the protobuf message (e.g. `PublishRelease`, `SetAccessPolicy`)
2. Signs it with the Ed25519 private key
3. Submits via GraphQL `sendTx` or ConnectRPC

Proto definitions live in the `proto/` directory of the Mojave repo and in the `ddex-proto` module for DDEX types. The `.proto` files are the contract — generate types for your language (`prost` for Rust, `protobuf-es` for TypeScript).

### Transaction types a client might submit

| Transaction | Who signs | What it does |
|-------------|----------|-------------|
| `PublishRelease` | Content owner | Claim ownership of uploaded content with DDEX metadata |
| `SetAccessPolicy` | Owner / admin | Set who can access content and under what terms |
| `AddPolicy` / `RemovePolicy` | Owner / admin | Add/remove Casbin policy rules |
| `DelegateRole` | Owner / admin | Grant admin/distributor role to another key |
| `TransferEntitlement` | Owner | Transfer ownership to another key |
| `RevokeEntitlement` | Owner / grantor | Revoke a role |
| `DeployScript` | Owner / admin | Deploy a Goja policy or proof script |
| `RegisterProof` | Owner / admin | Register a proof definition for attestations |

## Token

| | |
|---|---|
| Token | MOJ |
| Base unit | grain |
| Conversion | 1 MOJ = 1,000,000,000 grains (10^9) |

All on-chain values are in grains. Gas fees for normal operations (browsing, purchasing) are negligible. Storage fees (uploading) are proportional to file size × replication set size.

## Rust crate recommendations (Tauri)

| Concern | Crate | Notes |
|---------|-------|-------|
| BitTorrent | `librqbit` | Pure Rust BT client. Leech `.flac.tdf` and PNGs from validators. |
| Signing | `ed25519-dalek` | Ed25519 signing for transactions and access requests. |
| Encryption | `x25519-dalek` | X25519 ECDH for DEK wrapping/unwrapping. |
| Symmetric crypto | `aes-gcm` | AES-GCM for DEK-based decryption of TDF payloads. |
| Protobuf | `prost` | Generate Rust types from the same `.proto` files validators use. |
| HTTP | `reqwest` | For GraphQL queries, ConnectRPC calls, image/content downloads. |
| Keychain | `keyring` | Cross-platform secure storage for X25519 private key. |

## Browser considerations

- No native BitTorrent — download `.flac.tdf` via HTTP GET from the validator API.
- Storage: **OPFS** (Origin Private File System) — persistent, sandboxed, survives page reloads.
- Crypto: Web Crypto API for ECDH (X25519) and AES-GCM. `@noble/ed25519` for signing.
- Wallet: Phantom Connect or similar adapter for Ed25519 signing.
- WebTorrent (BitTorrent over WebRTC) can be added later as a P2P optimization.

## Metadata

Release metadata follows the **DDEX ERN** standard — the music industry's existing format for electronic release notification. Types are defined in protobuf (`ddex-proto` module) and exposed through GraphQL.

A release includes:
- Artist information
- Track titles, ISRCs
- Territory and rights information
- Genre, release date, label
- References to FLAC CID, encrypted CID, and image CIDs

This means metadata is already in a format that Warner, Spotify, Apple Music, and any distributor understands. Portability is built in.

## Roles and entitlements

Content has a role hierarchy:

| Role | Can do |
|------|--------|
| `owner` | Full control. Transfer, delegate, set policies. Created by `PublishRelease`. |
| `admin` | Manage policies, upload on behalf of owner. Cannot transfer. |
| `distributor` | Grant consumer access within delegated scope (territory, time). |
| `consumer` | Decrypt and play. |

Roles cascade: revoking an admin revokes everything that admin delegated.

## Access policy types

| Type | Backed by | Example |
|------|-----------|---------|
| `public` | Built-in | Anyone can access, no payment |
| `direct_grant` | Casbin ACL | Owner explicitly grants specific keys |
| `role_based` | Casbin RBAC | Any entity with `consumer` role |
| `conditional` | Casbin ABAC | Access if territory/time/tier match |
| `programmable` | Goja JS | Custom logic for complex licensing |
