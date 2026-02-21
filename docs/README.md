# Mojave Documentation

## Why Mojave exists

The music industry runs on intermediaries. Between the artist and the listener sits a stack of platforms, distributors, aggregators, and streaming services — each taking a cut, each controlling access, each holding data the artist can't fully see or verify. An independent artist uploading to Spotify doesn't own their distribution infrastructure. A label shipping to Apple Music trusts Apple's servers, Apple's DRM, Apple's analytics. The artist gets a dashboard someone else built and a check they can't independently verify.

This isn't a technology problem. The technology to distribute files, encrypt content, and track access has existed for decades. It's a trust problem. The infrastructure is owned by the platforms, and the platforms answer to shareholders, not artists.

Mojave replaces the platform with a protocol. No single entity owns the infrastructure — a network of elected, accountable validators does. No single entity controls access — cryptographic policies enforced by consensus do. No single entity holds the analytics — on-chain audit trails that anyone can query do.

The bet is simple: if you give artists and labels the tools to own their distribution — real ownership, with cryptographic proof, portable metadata, and transparent economics — they'll use them. Not because decentralization is a buzzword, but because the alternative is trusting a platform that can change its terms, delist your catalog, or shut down.

### What this enables that doesn't exist today

- **An independent artist** uploads a track, sets a price, and sells directly to fans. No distributor, no 30% platform cut, no 6-month payment delay. The sale is an on-chain transaction. The artist sees it in real time.
- **A label** manages a catalog across territories with structured access control — the same RBAC model they'd use with AWS IAM, but enforced by consensus instead of a cloud provider they don't control.
- **A distributor** can prove to a licensor that content was delivered N times in a specific territory, with a signed attestation from the network's validators — not a PDF report they generated themselves.
- **A fan** buys a track and actually owns it. The encrypted file sits on their disk. The DEK is wrapped to their device. They can play it offline, forever, without a subscription. If the network disappeared tomorrow, anyone who has the file and the DEK keeps their music.
- **A developer** builds a music player, a marketplace, or a recommendation engine on top of open chain state and a GraphQL API — no API keys, no rate limits, no terms of service that can be revoked.

### What this is not

- **Not a streaming service.** Mojave is infrastructure, not a product. There's no "Mojave app" for consumers (though anyone can build one). The protocol handles distribution, access control, and payment. The user experience is built on top by anyone.
- **Not DRM.** Once audio is decrypted and playing through speakers, it can be recorded. Mojave provides cryptographic access control and an audit trail, not copy protection. This is the same reality Spotify, Apple Music, and every other digital platform lives with.
- **Not a blockchain for blockchain's sake.** CometBFT is used because the problem requires consensus (who owns what, who can access what, which validators hold which keys) and a bounded trust set (elected validators, not anonymous miners). If a centralized server could be trusted to do this, you'd use a centralized server.

## How to read these docs

The docs are structured around concerns, not layers. Start with `architecture.md` for the full picture, then go deep on whichever concern matters to you.

### Reading order

**If you're an engineer building Mojave:**

1. [architecture.md](architecture.md) — the complete system design. Four planes, all flows, all transaction types.
2. [storage.md](storage.md) — PebbleDB key spaces (chain store + local store). This is what you implement.
3. [content.md](content.md) — on-disk file layout, BitTorrent integration, reconciliation loop. This is what the validator manages on the filesystem.
4. [economics.md](economics.md) — token model, fee structure. This is what the ABCI++ app enforces (CometBFT 1.x).
5. [governance.md](governance.md) — elections, takedowns, jurisdictional compliance. This is what the governance module implements.

**If you're an artist or label evaluating the system:**

1. [architecture.md](architecture.md) — focus on: Overview, Roles, Upload Flow, Access / Consumption Flow, Library Download.
2. [economics.md](economics.md) — focus on: Token, Content purchase fees, Transaction fee examples.
3. [governance.md](governance.md) — focus on: Validator elections (who runs the network), Takedowns (how copyright is enforced).

**If you're a developer building on top (music player, marketplace, etc.):**

1. [`PROTOCOL.md`](../PROTOCOL.md) — **start here**. The client-facing interface contract: auth, API, crypto, content access, payment, Rust/browser recommendations. This is everything you need without reading validator internals.
2. [architecture.md](architecture.md) — if you need deeper context: API Layer (GraphQL + ConnectRPC), Library Download, Keys (device-scoped encryption).
3. [content.md](content.md) — if you need BitTorrent integration details.
4. [economics.md](economics.md) — if you need payment mechanics beyond what PROTOCOL.md covers.

### Document map

| Document | What it covers | Key audience |
|----------|---------------|-------------|
| [`PROTOCOL.md`](../PROTOCOL.md) | Client interface contract — auth, API, crypto, content access, payment flows, Rust/browser crate recommendations. Self-contained; designed to be copied into client repos. | Client developers, LLMs in external repos |
| [architecture.md](architecture.md) | System overview, four planes, all actors, on-chain state, transaction types, upload/access/download flows, policy plane (Casbin + Goja), proofs & attestations, networking, API layer, design principles, trust assumptions | Everyone |
| [storage.md](storage.md) | Two PebbleDB stores — chain store (consensus state, key spaces, secondary indexes) and local store (validator DEKs, processing scratch, sync state). Rebuilding from peers. | Engineers |
| [content.md](content.md) | On-disk file layout (`.flac.tdf` + `.png`), directory sharding, `gocloud.dev` integration, BitTorrent integration (seeding, leeching, good samaritans, dead seed problem), reconciliation loop, lifecycle, disk sizing | Engineers, validator operators |
| [economics.md](economics.md) | Native token (MOJ/grains), gas fees, storage fees, content purchases, validator rewards, staking, genesis allocation, inflation schedule, bootstrapping phases, fee examples, governance parameters | Everyone |
| [governance.md](governance.md) | Validator elections (social + staking), oracle elections (per-network), copyright takedowns (DEK removal), counter-notices, jurisdictional compliance, content flagging, governance proposals | Everyone |

### Diagrams

All diagrams live in [diagrams/](diagrams/) as `.mermaid` files and are referenced inline from the docs they support.

| Diagram | In | Shows |
|---------|------|-------|
| [overview.mermaid](diagrams/overview.mermaid) | architecture.md | Four-plane architecture, how all components connect |
| [upload.mermaid](diagrams/upload.mermaid) | architecture.md | Full upload flow — client to validator to chain to peers |
| [access-direct.mermaid](diagrams/access-direct.mermaid) | architecture.md | Consumer requests access from a DEK holder validator |
| [access-forwarded.mermaid](diagrams/access-forwarded.mermaid) | architecture.md | Consumer hits replication-only validator, redirected to DEK holder |
| [library-download.mermaid](diagrams/library-download.mermaid) | architecture.md | iTunes-style library sync — query, download, play |
| [policy-evaluation.mermaid](diagrams/policy-evaluation.mermaid) | architecture.md | Decision flowchart — public / Casbin / Goja → grant or deny |
| [attestation.mermaid](diagrams/attestation.mermaid) | architecture.md | Proof script execution and signed attestation delivery |
| [keys.mermaid](diagrams/keys.mermaid) | architecture.md | Ed25519 identity, device X25519, recovery key derivation, ECDH wrapping |
| [dek-distribution.mermaid](diagrams/dek-distribution.mermaid) | architecture.md | Where wrapped DEKs live — on-chain, off-chain, device-scoped |
| [dek-recovery.mermaid](diagrams/dek-recovery.mermaid) | architecture.md | Uploader recovers DEK holders after full churn |
| [roles.mermaid](diagrams/roles.mermaid) | architecture.md | Role hierarchy — owner, admin, distributor, consumer |
| [networking.mermaid](diagrams/networking.mermaid) | architecture.md | Three communication layers — reactors, BitTorrent, API |
| [type-generation.mermaid](diagrams/type-generation.mermaid) | architecture.md | Proto → Go → GraphQL pipeline across ddex-proto and mojave repos |
| [validator-churn.mermaid](diagrams/validator-churn.mermaid) | architecture.md | Validator departure — unbonding, DEK holder replacement, optional rotation |
| [governance.mermaid](diagrams/governance.mermaid) | governance.md | Election process, validator/oracle sets, recall mechanism |
| [takedown.mermaid](diagrams/takedown.mermaid) | governance.md | Takedown flow — claim, review, counter-notice, DEK removal |

### Design decisions (token and payments)

- **MOJ for everything.** One native token (MOJ) for gas, storage fees, content purchases, staking, and validator rewards. No attestation-based USDC or multi-currency settlement in the protocol for now; keeps implementation simple. See [economics.md](economics.md).
- **Per-chain MOJ.** Each chain has its own MOJ economy; native tokens are not transferable across chains (no IBC). Liquidity and “get in/out in USD” can be provided by exchanges listing MOJ; that is exchange-layer.
- **Faucet for bootstrap.** Faucet (rate-limited) is the primary way to get MOJ into users' hands until exchange listings or other on-ramps exist. Validators do not sell MOJ for USDC; that stays off the protocol.
- **USD in UX only.** People think in USD; clients display USD equivalent (price feed) and can support “price in USD” with conversion to MOJ at purchase. No on-chain USDC or EigenLayer/ETH lock-in.

## Open questions and known gaps

These are unresolved design questions identified during documentation review. They don't block implementation but need answers before the design is final.

### Protocol design

**1. Who pays storage fees?**
`UploadComplete` is signed by the validator, but economics.md says the client's account is debited for storage fees. The client hasn't signed this transaction. Either the client needs to pre-fund or co-sign the upload, or the fee is deducted from a pre-authorized escrow. This needs a concrete mechanism.

**2. How does `public` policy work with DEK wrapping?**
The policy types table says public content has the "DEK published in the clear." But the entire encryption model assumes wrapped DEKs. Does `public` mean the DEK is literally stored unencrypted on-chain? Or does it mean the validator auto-grants to any requester without checking identity? The latter is simpler and consistent with the wrapping model.

**3. Where are validator X25519 public keys registered?**
Consumers use device-scoped keys (not on-chain). Uploaders derive recovery keys (on-chain only as the wrapped DEK, not the key itself). But validators need their X25519 public keys known so the upload validator can wrap DEKs for them. Are these registered on-chain as part of the validator record? The storage.md validator key space only has Ed25519.

**4. Multi-track releases and album-level policies.**
All state is keyed by individual FLAC CID. An album with 12 tracks is 12 separate uploads, 12 separate policies, 12 separate access grants. Is there a grouping mechanism (a "release CID" that bundles tracks)? Or does the client manage this by setting the same policy on all tracks? The DDEX ERN supports multi-track releases, but the on-chain model is per-track.

**5. How is `ReportPlay` structured?**
The proofs section mentions a `ReportPlay` transaction for voluntary play count reporting, but it's not listed in the transaction types and has no schema. If it exists, what prevents spam? (A consumer could report a million plays.) If it's opt-in and trust-the-reporter, is it useful enough to include?

**6. How is consumer territory determined?**
The Casbin ABAC model and Goja scripts reference `request.territory` for territorial access control. Is this self-reported by the consumer? Inferred from IP by the validator? If self-reported, it's trivially spoofable. If IP-based, it contradicts the wallet-only authentication model.

### Implementation details

**7. CID format and hash algorithm.**
The docs use "CID" everywhere but never specify the hash algorithm (SHA-256, BLAKE3, etc.) or encoding (multihash, raw hex, base58). This affects content addressing, BitTorrent rendezvous, directory sharding, and cross-system interoperability.

**8. BitTorrent infohash derivation.**
`content.md` calls `deriveInfohash(cid)` but doesn't specify the derivation. BitTorrent v1 infohashes are SHA-1 of torrent metadata (not content hashes). BitTorrent v2 uses SHA-256 Merkle trees. How does CID → infohash work? Does each content file get a `.torrent` metafile, or is the DHT used with a derived key?

**9. Upload validator selection.**
The upload flow starts with "Client sends raw audio to Upload Validator" but never specifies how the client picks which validator. Random selection? Closest by latency? Explicit choice? This affects availability (what if the chosen validator is overloaded?) and trust (the upload validator sees unencrypted audio).

**10. Account balance tracking.**
The accounts key space in storage.md has `pubkey`, `nonce`, `created_at` but no `balance` field. The economics doc describes payments and transfers. Where do balances live?

### Governance gaps

**11. Epoch boundaries.**
governance.md says validators are "added to the active set at the next epoch boundary." Epochs are never defined. CometBFT 1.x uses ABCI++ (FinalizeBlock, not legacy EndBlock); validator set changes are driven by the application in the ABCI++ lifecycle — is that the epoch, or is there a higher-level concept?

**12. Pending takedowns on oracle recall.**
If an oracle is recalled while they have a pending `TakedownRequest` (in the counter-notice window), what happens? Is the takedown auto-dismissed? Does another oracle take over? The doc is silent on this.

**13. Transaction types from governance.md not in architecture.md.**
The governance doc introduces new transaction types (`SubmitCandidacy`, `RecallValidator`, `OracleCandidacy`, `RecallOracle`, `TakedownRequest`, `TakedownResolve`, `CounterNotice`, `FlagContent`, `UnflagContent`) that aren't listed in architecture.md's Transaction Types section.

### Economics

**14. One-time storage fees vs. indefinite storage.**
Storage fees are one-time at upload, but validators store and seed indefinitely. Over years, the validator's ongoing costs (disk, bandwidth) may exceed the one-time payment. Is there a mechanism for ongoing compensation? Or do block rewards and gas fees cover the long-term cost? This should be explicit.

**15. Payment atomicity in GrantAccess.**
economics.md says content purchases transfer MOJ from consumer to content owner atomically within the `GrantAccess` transaction. But `GrantAccess` is "signed by validator" per architecture.md. How does the validator transfer the consumer's funds? The consumer needs to authorize the payment somehow — pre-signed payment, escrow, or a two-step process.

### Content

**16. Good samaritan nodes after takedowns.**
After a takedown, validators delete DEKs and optionally stop seeding. But good samaritans and consumers who already have the `.flac.tdf` can keep seeding undecryptable ciphertext indefinitely. This is probably acceptable (it's useless ciphertext), but should be stated explicitly as a known property.

**17. DDEX ERN validation.**
Who validates that the DDEX metadata in a `PublishRelease` transaction is well-formed? The upload validator during processing? The ABCI++ application during transaction validation (e.g. in ProcessProposal or FinalizeBlock)? Peer validators during consensus? If nobody validates it, malformed metadata goes on-chain permanently.
