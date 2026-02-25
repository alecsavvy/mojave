# Mojave Documentation

**Mojave is your decentralized music library:** rip your vinyl and CDs, import your digital copies, and buy directly from artists. One place to own your music — encrypted, portable, offline-first. In the future: digital trades and secondhand sales. A legal, better alternative to LimeWire meets iTunes, with real ownership and payments to rights holders.

## Why Mojave exists

**For you (the listener):** Your music is scattered — streaming services that can vanish, downloads locked to one app, physical media that isn’t in the cloud. Mojave is a single library you control: digitize what you own (rips, imports), buy from artists who sell here, and eventually trade or resell. No platform can delete your collection or change the rules after you’ve paid.

**For artists and labels:** The music industry runs on intermediaries. Between the artist and the listener sits a stack of platforms, distributors, and streaming services — each taking a cut, each controlling access. Mojave replaces the platform with a protocol. No single entity owns the infrastructure; a network of elected, accountable validators does. No single entity controls access; cryptographic policies enforced by consensus do. The bet: give artists and labels the tools to own their distribution, and give listeners one place to own their library — rips, purchases, and (future) resales — and they’ll use it.

### What this enables that doesn't exist today

- **You** digitize your personal library — rip vinyl, CDs, or import digital copies — and store them in the same system where you buy music. One library: your rips, your purchases, your keys. Offline playback, no subscription.
- **You** buy a track from an artist or seller and actually own it. The encrypted file sits on your disk; the DEK is wrapped to your device. If the network disappeared tomorrow, you keep your music.
- **An independent artist** uploads a track, sets a price, and sells directly to fans. No distributor, no 30% platform cut. The sale is an on-chain transaction.
- **A label** manages a catalog across territories with structured access control — the same RBAC model they’d use with AWS IAM, but enforced by consensus.
- **A distributor** can prove to a licensor that content was delivered N times in a specific territory, with a signed attestation from the network’s validators.
- **In the future:** digital trades and secondhand sales — transfer or resell your copy under rules set by rights holders and the protocol.
- **A developer** builds a music player, a marketplace, or a recommendation engine on top of open chain state and a GraphQL API — no API keys, no rate limits, no revocable terms of service.

### Showcase frontends (mojave.audio)

To demonstrate that the protocol is infrastructure and the UX is unbounded, the plan is to run multiple frontends on [mojave.audio](https://mojave.audio):

| Frontend | Purpose |
|----------|---------|
| **mojave.audio** | Main site — reference UI; aesthetic: chill vinyl listening room + late-2000s metallic desktop (see [design.md](design.md)) |
| **spotify.mojave.audio** | Spotify-style experience (browse, playlists, discovery) |
| **bandcamp.mojave.audio** | Bandcamp-style (artist-first, pay what you want, ownership) |
| **audius.mojave.audio** | Audius-style (community, trending, social) |
| **itunes.mojave.audio** | iTunes-style (library, store, own your music) |
| **myspace.mojave.audio** | MySpace-style (artist profiles, top friends, custom pages, early social music) |

Same protocol, same chain; different UX paradigms. The **Tauri desktop app** is the reference client: your library, playback, and optional seeding in one place — so you're not "I bought it on Bandcamp, now I open VLC."

### What this is not

- **Not a streaming service.** Mojave is infrastructure; the reference experience is the Tauri app (library + player + sync). The protocol handles distribution, access control, and payment. Anyone can build other clients on top.
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

**If you care about design and aesthetic:**

1. [design.md](design.md) — Vinyl listening room + late-2000s metallic desktop; character, palette, showcase frontends.

**If you care about the "personal library" reorientation (rips, imports, future resale):**

1. [personal-library-vision.md](personal-library-vision.md) — **start here**. Vision, catalog vs personal library, rip/import flow, future trades/resale, open questions.
2. [architecture.md](architecture.md) — Personal library section; how it fits with existing upload/access flows.
3. [storage.md](storage.md) — (When added) library key spaces and content source.

**If you're a developer building on top (music player, marketplace, etc.):**

1. [`PROTOCOL.md`](../PROTOCOL.md) — **start here**. The client-facing interface contract: auth, API, crypto, content access, payment, Rust/browser recommendations. This is everything you need without reading validator internals.
2. [architecture.md](architecture.md) — if you need deeper context: API Layer (GraphQL + ConnectRPC), Library Download, Keys (device-scoped encryption), Personal library.
3. [content.md](content.md) — if you need BitTorrent integration details.
4. [economics.md](economics.md) — if you need payment mechanics beyond what PROTOCOL.md covers.

### Document map

| Document | What it covers | Key audience |
|----------|---------------|-------------|
| [design.md](design.md) | **Aesthetic:** vinyl listening room + late-2000s metallic desktop; character, palette, showcase frontends. | Design, frontend |
| [personal-library-vision.md](personal-library-vision.md) | **Reorientation:** decentralized personal music library — rip/import, purchase, future trades/resale. Catalog vs personal library, flows, open questions. | Everyone (vision and product) |
| [`PROTOCOL.md`](../PROTOCOL.md) | Client interface contract — auth, API, crypto, content access, payment flows, Rust/browser crate recommendations. Self-contained; designed to be copied into client repos. | Client developers, LLMs in external repos |
| [architecture.md](architecture.md) | System overview, four planes, all actors, on-chain state, transaction types, upload/access/download flows, policy plane (entitlement-first; Casbin + Goja optional for fine-grained), proofs & attestations, networking, API layer, design principles, trust assumptions; **personal library** (catalog vs library, rip/import). | Everyone |
| [storage.md](storage.md) | Two PebbleDB stores — chain store (consensus state, key spaces, secondary indexes) and local store (validator DEKs, processing scratch, sync state). Rebuilding from peers. | Engineers |
| [content.md](content.md) | On-disk file layout (`.flac.tdf` + `.png`), directory sharding, `gocloud.dev` integration, BitTorrent integration (seeding, leeching, good samaritans, dead seed problem), reconciliation loop, lifecycle, disk sizing | Engineers, validator operators |
| [economics.md](economics.md) | No native token; USDC attestations for subscriptions (library size) and content purchases; artists as distributors; validators take a cut; users and good samaritans seed | Everyone |
| [governance.md](governance.md) | Validator elections (social + staking), publisher groups (states), per-group takedown authority, copyright takedowns (DEK removal), counter-notices, jurisdictional compliance, content flagging, governance proposals | Everyone |

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
| [governance.mermaid](diagrams/governance.mermaid) | governance.md | Election process, validator set, publisher groups, recall mechanism |
| [takedown.mermaid](diagrams/takedown.mermaid) | governance.md | Takedown flow — claim, review, counter-notice, DEK removal |

### Design decisions (payments)

- **No native token (no MOJ).** Everyone has an Ed25519 pubkey. Payments use **USDC attestations** (user subscriptions by library size + content purchases). See [economics.md](economics.md).
- **Artists as distributors; validators take a cut.** Artists grant access on USDC purchase attestation; validators take a cut; artist sale volume + user subscription fees pay for hosting. Users and good samaritans also seed (BitTorrent). Liquidity and “get in/out in USD” can be provided by exchanges listing MOJ; that is exchange-layer.
- (price feed) and can support “price in USD” with conversion to MOJ at purchase. No on-chain USDC or EigenLayer/ETH lock-in.

## Open questions and known gaps

These are unresolved design questions identified during documentation review. They don't block implementation but need answers before the design is final.

### Protocol design

**1. Payment for replication (deferred).**
Who pays validators for replication and how is deferred. The assumed direction is a **USDC subscription** based on amount of content (e.g. $/month per GB or per tracks). Concrete mechanism (on-chain vs off-chain, tiers, metering) is TBD. See [economics.md](economics.md).

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

**10. Accounts.** With no native token, accounts may only need `pubkey` and `nonce` (replay protection). Payments are via USDC attestations (off-chain or on the chain where USDC lives); no on-chain balance in the protocol.

### Governance gaps

**11. Epoch boundaries.**
governance.md says validators are "added to the active set at the next epoch boundary." Epochs are never defined. CometBFT 1.x uses ABCI++ (FinalizeBlock, not legacy EndBlock); validator set changes are driven by the application in the ABCI++ lifecycle — is that the epoch, or is there a higher-level concept?

**12. Pending takedowns when a group’s takedown authority changes.**
If a group updates its designated takedown authority (e.g. `SetGroupTakedownAuthority`) while a takedown for that group’s content is in the counter-notice window, what happens? Does the original authority’s pending request stand, or does the new authority take over? The doc is silent on this.

**13. Transaction types from governance.md not in architecture.md.**
The governance doc introduces new transaction types (`SubmitCandidacy`, `RecallValidator`, `OracleCandidacy`, `RecallOracle`, `TakedownRequest`, `TakedownResolve`, `CounterNotice`, `FlagContent`, `UnflagContent`) that aren't listed in architecture.md's Transaction Types section.

### Economics

**14. Replication payment.** Users pay a **USDC subscription** (by library size) via attestation; that revenue pays validators for user library backup. Artist sale volume (validators take a cut) pays for artist hosting. See [economics.md](economics.md).

**15. Purchase attestation flow.** Content purchases use USDC attestations: user pays in USDC; attestation proves payment; artist (distributor) or validator verifies and grants access. Attestation format and verification (on-chain vs off-chain) are TBD. Validators take a cut of the sale.

### Content

**16. Good samaritan nodes after takedowns.**
After a takedown, validators delete DEKs and optionally stop seeding. But good samaritans and consumers who already have the `.flac.tdf` can keep seeding undecryptable ciphertext indefinitely. This is probably acceptable (it's useless ciphertext), but should be stated explicitly as a known property.

**17. DDEX ERN validation.**
Who validates that the DDEX metadata in a `PublishRelease` transaction is well-formed? The upload validator during processing? The ABCI++ application during transaction validation (e.g. in ProcessProposal or FinalizeBlock)? Peer validators during consensus? If nobody validates it, malformed metadata goes on-chain permanently.

### Personal library (reorientation)

**18. Personal library: client-side vs validator-side encryption.**
For rip/import (digitizing your own collection), should the client encrypt locally so validators never see plaintext, or reuse the current upload pipeline (validator transcodes and encrypts)? Client-side maximizes privacy; validator-side reuses implementation. See [personal-library-vision.md](personal-library-vision.md).

**19. Transfer of consumer entitlement (resale / trade).**
Secondhand sales and digital trades require transferring a consumer entitlement from one key to another. No transaction type exists yet. Need: transferability rules (who can transfer, who gets a cut), atomicity, and policy/DDEX hooks. See [personal-library-vision.md](personal-library-vision.md).
