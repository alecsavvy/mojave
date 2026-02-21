# Mojave Architecture

## Overview

> **[System overview diagram](diagrams/overview.mermaid)**

Mojave is a decentralized music distribution system built on four planes:

- **Consensus plane** (CometBFT): the source of truth. Stores state, orders transactions, coordinates validators via ABCI events. Nothing is "real" until it's on-chain.
- **Storage plane** (BitTorrent + gocloud.dev): encrypted file replication. BitTorrent is the replication protocol between validators. gocloud.dev is each validator's local storage abstraction — it lets them back their blob store with disk, S3, GCS, whatever. These are complementary: gocloud.dev is "where I keep my copy," BitTorrent is "how I get copies to/from peers."
- **Encryption plane** (OpenTDF): cryptographic enforcement of access. OpenTDF manages DEK/KEK wrapping so that only authorized keyholders can decrypt content. The chain replaces the centralized Key Access Server (KAS) — wrapped DEKs live on-chain as state, not held by any single party.
- **Policy plane** (Casbin + Goja): authorization logic and programmable business rules. Casbin provides structured IAM — roles, groups, hierarchies, RBAC/ABAC — backed by a custom adapter that reads/writes policy state on-chain. A sandboxed JavaScript runtime (Goja) handles logic that can't be expressed as Casbin policies — complex licensing, custom attestations, dynamic conditions. This is where DDEX's gap lives: DDEX assumes bucket/web-service-level access control, which doesn't exist here, so the policy plane fills that role explicitly.

## Keys

> **[Key derivation and wrapping diagram](diagrams/keys.mermaid)**

All actors use Ed25519 keypairs for identity (CometBFT default). Ed25519 is a signature scheme — it can't be used for encryption directly.

- **Ed25519** (Edwards form): the identity key. Used for signing — transactions, entitlements, attestations, access requests.
- **X25519** (Montgomery form): device-local encryption keys. Used for DEK wrapping via ECDH. Generated per-device, never on-chain.

### Device-scoped encryption keys

Each device generates its own X25519 keypair locally and stores it in the device keychain / secure storage. Encryption keys are **not registered on-chain** — they exist only on the device. This gives natural multi-device support without any on-chain key management.

When a consumer requests access:

1. The device includes its X25519 public key in the access request.
2. The user signs the request with their Ed25519 wallet key (proving identity).
3. The validator verifies the signature, checks the Casbin policy.
4. If authorized: wraps the DEK against the device's X25519 public key, returns it directly.
5. The device unwraps the DEK with its local X25519 private key and decrypts the content.

Properties:

- **Multi-device is trivial.** Each device requests its own wrapped DEK. Phone, laptop, tablet — each gets a DEK wrapped to its own key.
- **Stolen wrapped DEKs are useless.** A wrapped DEK from one device can't be unwrapped on another — different X25519 key.
- **No encryption key state on-chain.** The chain only knows Ed25519 signing keys. Device encryption keys are ephemeral from the chain's perspective.
- **Works with any wallet.** Social login (Phantom, Google, Apple) or self-custody — the wallet only needs to sign. No private key access required for encryption.

### Uploader recovery key

The uploader's on-chain wrapped DEK exists as a last-resort recovery mechanism — it's only needed if every validator in the DEK holder set goes down simultaneously. With a well-chosen set this is extremely unlikely, and even then the encrypted `.tdf` blob is still in the replication set — the content isn't lost, just the ability to wrap new DEKs.

This is not a core protocol concern. It can be solved at a layer above the protocol:

- **Self-custody users**: derive X25519 from Ed25519 directly (birational map via `filippo.io/edwards25519`). Store the recovery wrapped DEK on-chain.
- **Custodial recovery service**: a third-party or L2 service holds a recovery key on behalf of social login users. The uploader's wrapped DEK is wrapped against the service's key.
- **Multisig / social recovery**: split the recovery key across trusted parties (e.g. the user + 2 of 3 trusted contacts).

All other DEK wrapping (consumer access, validator distribution) uses device-scoped keys and doesn't depend on the recovery key.

### DEK wrapping flow

Generate an ephemeral X25519 keypair, ECDH with the recipient's X25519 public key (device key or recovery key), derive a symmetric wrapping key, encrypt the DEK. The ephemeral public key is stored alongside the wrapped DEK so the recipient can perform their half of the ECDH.

## Actors

- **Clients**: music creators / rights holders. Have Ed25519 keypairs. Upload content, assert ownership, manage access.
- **Validators**: run CometBFT nodes. Have Ed25519 keypairs (CometBFT validator keys). Process uploads, enforce policies, replicate files, serve content.
- **Consumers**: want to access / stream music. Have Ed25519 keypairs. Granted access through on-chain policies.

## On-Chain State Machine

> See [storage.md](storage.md) for the full PebbleDB key space layout (chain store + local store) and [content.md](content.md) for the on-disk file layout.

The CometBFT application manages these core state objects:

### Uploads (keyed by FLAC CID)

- FLAC CID (content hash of the unencrypted transcoded audio — the canonical content identifier)
- Encrypted CID (content hash of the encrypted `.tdf` blob)
- Image CIDs (content hashes of associated artwork, normalized to PNG)
- Uploader's public key
- Status: `processing_complete` | `published`
- Timestamp
- **Replication set**: which validators store and seed the encrypted `.tdf` blob and images. Controls availability.
- **DEK holder set**: which validators hold wrapped DEKs (in their local store) and can issue `GrantAccess`. Controls security blast radius.

**Audio** is encrypted and stored as a `.tdf` blob — a standard OpenTDF container holding the encrypted payload. Validators in the replication set discover and seed the blob via BitTorrent using the encrypted CID as the rendezvous point. Validators in the DEK holder set additionally hold the wrapped DEK in their local store, fetched from peers via p2p.

**Images** (cover art, artist photos, etc.) are normalized to PNG and stored **unencrypted**. They are content-addressed by CID, replicated across the replication set, and served directly by any validator. No DEK, no access control — images are public by design. This allows any UI built on top of a validator or RPC endpoint to display artwork, build browse/search interfaces, and render release pages without requiring access grants. Audio is the protected asset; images are the public-facing layer.

The two content types follow the same normalization philosophy: all audio → FLAC, all images → PNG. Predictable formats mean clients don't need format negotiation.

These two sets are independently configurable by the uploader. Encrypted `.tdf` blobs are ciphertext — safe to replicate broadly for availability. DEKs are the sensitive part — distributing them narrowly limits exposure if a validator departs.

Example configurations:

| Profile | Replication | DEK holders | Tradeoff |
|---------|-------------|-------------|----------|
| Maximum availability | 60/60 validators | 60/60 validators | Any validator can serve the file and grant access. Largest blast radius on validator departure. |
| Controlled access | 60/60 validators | 5/60 validators + uploader | File is highly available, but only a small trusted set can grant access. Validator departure blast radius is 1/5. DEK rotation is cheap. |
| Lightweight | 5/60 validators | 5/60 validators | Low storage footprint. Sufficient for niche releases or limited-audience content. |

When a consumer requests access from a validator that is in the replication set but not the DEK holder set, that validator forwards or redirects the request to a DEK-holding validator. The consumer still gets the encrypted file from whichever validator is closest/fastest (BitTorrent handles this naturally), but the `GrantAccess` operation is handled by a DEK holder.

### Wrapped DEKs

> **[DEK distribution diagram](diagrams/dek-distribution.mermaid)**

Wrapped DEKs are the DEK encrypted against a specific entity's X25519 public key. They are ciphertext — unwrapping requires the corresponding X25519 private key. Never raw DEKs.

**On-chain (keyed by CID):**

- **The uploader's wrapped DEK is always on-chain.** This is the only wrapped DEK that lives in chain state. Wrapped against the uploader's recovery key (derived from their Ed25519 key — see Keys section). It is the root of trust and recovery mechanism of last resort. If every validator in the DEK holder set departs, the uploader can come online, unwrap their own DEK, re-wrap it for a new set of validators via `RecoverDEKHolders`, and restore access.

**Off-chain (validator local store, distributed via CometBFT reactors):**

- Validator wrapped DEKs are stored locally by each validator. The chain's DEK holder set configuration determines which validators *should* have a wrapped DEK for a given CID — the actual wrapped DEK data is distributed asynchronously via p2p.
- At upload time, the upload validator wraps the DEK for each validator in the DEK holder set and sends the wrapped DEKs to them via CometBFT reactors.
- A new validator syncing the network downloads its wrapped DEKs from peers that already hold the raw DEK (existing DEK holders or the uploader).

**Consumer wrapped DEKs (device-scoped, not persisted):**

- Not stored on-chain or on validators. Wrapped at access time against the requesting device's X25519 public key and returned directly to the device.
- Each device has its own X25519 keypair. A wrapped DEK from one device is useless on another.
- **Expiration** (optional): can carry an expiry (block height or timestamp). After expiry, the device requests a fresh wrapped DEK from a validator. The validator re-checks the Casbin policy at each request, so revocation takes effect at the next request. Permanent grants (purchased content) can omit the expiry.
- The on-chain `GrantAccess` record logs the access event (Ed25519 key + CID + timestamp) for the proof/attestation system, but contains no wrapped DEK data.

### Releases (keyed by FLAC CID)

- Full DDEX ERN
- Signed by the client's key (this *is* the entitlement)
- References both FLAC CID and encrypted CID

### Entitlements (keyed by (CID, public key))

On-chain objects representing rights over content. An entitlement says "public key X has role Y over content Z." They are the policy plane's primary state primitive.

- **Transferable**: an owner can transfer their entitlement (sell a catalog).
- **Delegable**: an owner can grant sub-entitlements (label delegates distribution rights to a partner).
- **Scopable**: entitlements can be constrained by territory, time window, or custom conditions.

### Roles

> **[Role hierarchy diagram](diagrams/roles.mermaid)**

A fixed set of role types that entitlements can carry:

- `owner` — created by `PublishRelease`, bound to the signer's key. Full control. Can transfer, delegate, set policies.
- `admin` — granted by owner. Can manage policies and upload on behalf of the owner. Cannot transfer ownership.
- `distributor` — granted by owner or admin. Can issue consumer access within their delegated scope (e.g. territory, time window).
- `consumer` — granted by a distributor (or directly by owner). Can decrypt and play.

For a label:

```
Label Key (owner of catalog)
  ├─ A&R Key (admin — can upload on behalf of artists)
  ├─ Distribution Partner Key (distributor — can grant consumer access in NA)
  └─ Artist Key (owner of individual tracks, royalty entitlement)
```

For an independent:

```
Artist Key (owner)
  └─ (directly grants consumer access, or sets a "public" policy)
```

Same model, different depth. The independent just doesn't use delegation.

### Access Policies (keyed by CID)

Declarative rules attached to content by its owner. Evaluated by validators before issuing wrapped DEKs.

- **Type**: `public` | `direct_grant` | `role_based` | `conditional` | `programmable`
- **Casbin model reference** (for `direct_grant`, `role_based`, `conditional`): references the on-chain Casbin model and policy set used for evaluation. A default model is available for simple cases.
- **Script CID** (for `programmable`): reference to an on-chain JavaScript policy function evaluated by Goja. Used when Casbin can't express the logic needed.

### Casbin Policies (keyed by model ID)

The Casbin adapter stores policy rules as structured on-chain state:

- **Model definition**: the PERM (Policy, Effect, Request, Matchers) model.
- **Policy rules**: `p` lines — e.g. `(sub, obj, act, eft)` tuples.
- **Group/role assignments**: `g` lines — e.g. `(user, role)` or `(user, role, domain)` tuples.
- **Scoped by content owner**: each owner's policies are isolated. A label's policy rules don't affect another label's content.

## Transaction Types

### Content lifecycle

1. **`UploadComplete`** (signed by validator) — "I processed this file. Here are the CIDs, the info hash, and wrapped DEKs for every validator."
2. **`PublishRelease`** (signed by client) — "I am the rights holder. Here is my DDEX ERN referencing these CIDs." This is the entitlement claim. The client's signature binds their identity to the content cryptographically.

### Policy management

3. **`SetAccessPolicy`** (signed by owner/admin) — "Here's who can access my content and under what terms." References a Casbin model + policies, or a Goja script, or both.
4. **`SetPolicyModel`** (signed by owner/admin) — "Here's the Casbin model for my content." Stores the PERM model definition on-chain. A default model is provided for simple use cases; labels and enterprises can define richer models with domain-scoped roles and ABAC attributes.
5. **`AddPolicy`** (signed by owner/admin) — "Add a Casbin policy rule." e.g. `p, alice_pubkey, track_cid, play` or `g, bob_pubkey, label-abc-staff`. Written to chain state via the Casbin adapter.
6. **`RemovePolicy`** (signed by owner/admin) — "Remove a Casbin policy rule."
7. **`DelegateRole`** (signed by owner/admin) — "Grant public key X the role of admin/distributor over CID Y with scope Z." Creates an on-chain entitlement. The delegator must hold a role with sufficient authority.
8. **`TransferEntitlement`** (signed by current owner) — "Transfer my owner entitlement to public key X." Catalog sales, rights transfers.
9. **`RevokeEntitlement`** (signed by grantor or owner) — "Remove public key X's entitlement." Cascades: revoking an admin revokes all roles that admin delegated.
10. **`DeployScript`** (signed by owner/admin) — "Store this JavaScript function on-chain." The script is content-addressed (CID) and can be referenced by access policies (for programmable policy evaluation), Casbin custom matchers, or proof definitions (for attestation generation). Validators must be able to execute it deterministically.

### Access

11. **`GrantAccess`** (signed by validator) — "I've verified this consumer is authorized per the on-chain policy." Recorded on-chain as an audit trail (Ed25519 signing key + CID + timestamp) for the proof/attestation system. The actual wrapped DEK is delivered directly to the requesting device (wrapped against the device's X25519 public key) and is not stored on-chain.
12. **`RevokeAccess`** (signed by owner/admin) — "Remove this grant." Validators will stop issuing new wrapped DEKs for this identity. Wrapped DEKs already on devices still work until they expire, but the device can't get a fresh one.

### Proofs

13. **`RegisterProof`** (signed by owner/admin) — "Register a proof definition for my content." References a deployed Goja script (by CID) that defines what the proof computes and the shape of the attestation it produces. The proof definition is on-chain so anyone can audit what's being proved.
14. **`SubmitAttestation`** (signed by validator) — "I ran proof script X against chain state at block height N and produced this result." The attestation is stored on-chain with the validator's signature. Optional — attestations can also be delivered directly to the requester off-chain if on-chain storage isn't needed.

### Validator lifecycle

> **[DEK recovery flow diagram](diagrams/dek-recovery.mermaid)** · **[Validator churn diagram](diagrams/validator-churn.mermaid)**

15. **`UpdateDEKHolderSet`** (signed by owner/admin) — "Change which validators should hold wrapped DEKs for my content." Updates the on-chain DEK holder set configuration. Validators newly added to the set will fetch their wrapped DEK from existing holders via p2p. Validators removed from the set should delete their local wrapped DEK (enforced by peer protocol, not cryptographically — they already have the key material).
16. **`RotateDEK`** (signed by owner/admin) — "Re-encrypt my content with a fresh DEK." A DEK-holding validator re-encrypts the file, generates a new uploader wrapped DEK for on-chain storage, distributes new wrapped DEKs to the current DEK holder set via p2p, re-seeds the new `.tdf` blob, and submits updated CIDs and uploader wrapped DEK on-chain. Cryptographically revokes access for any party not in the new wrapped DEK set.
17. **`RecoverDEKHolders`** (signed by uploader) — "Re-wrap my DEK for a new set of validators." The uploader unwraps their own DEK (using their private key against their always-on-chain wrapped DEK), wraps it for a new DEK holder set, and distributes the wrapped DEKs via p2p. The on-chain DEK holder set configuration is updated. Used when the original DEK holder set has degraded or been fully lost to churn. The uploader must be online, but no re-encryption is needed — the DEK and `.tdf` blob remain the same.

## Upload Flow

> **[Upload flow diagram](diagrams/upload.mermaid)**

The client specifies a **replication set** (which validators store the encrypted file) and a **DEK holder set** (which validators receive wrapped DEKs and can grant access).

```
Client                    Upload Validator           Chain              Peer Validators
  |                              |                     |                       |
  |-- raw audio + images         |                     |                       |
  |   + DDEX metadata             |                     |                       |
  |   + replication set config   |                     |                       |
  |   + DEK holder set config -->|                     |                       |
  |                              |                     |                       |
  |                    transcode audio to FLAC          |                       |
  |                    normalize images to PNG          |                       |
  |                    generate FLAC CID               |                       |
  |                    generate image CIDs             |                       |
  |                    generate DEK                    |                       |
  |                    encrypt FLAC with DEK → .tdf    |                       |
  |                    generate encrypted CID          |                       |
  |                    wrap DEK for client PK          |                       |
  |                    begin seeding .tdf (BitTorrent) |                       |
  |                    begin seeding PNGs (BitTorrent) |                       |
  |                              |                     |                       |
  |                              |-- UploadComplete tx ->|                       |
  |                              |   (FLAC CID,         |                       |
  |                              |    encrypted CID,    |                       |
  |                              |    image CIDs,       |                       |
  |                              |    client wrapped    |                       |
  |                              |    DEK, replication  |                       |
  |                              |    set, DEK holder   |                       |
  |                              |    set)              |                       |
  |                              |                     |-- upload.complete event ->|
  |                              |                     |                       |
  |<-- CIDs + metadata ---------|                     |  replication set:
  |                              |                     |    fetch .tdf via BitTorrent
  |                              |                     |    (rendezvous on encrypted CID)
  |                              |                     |    fetch PNGs via BitTorrent
  |                              |                     |    (rendezvous on image CIDs)
  |                              |                     |    store via gocloud.dev
  |                              |                     |  DEK holder set:
  |                              |                     |    fetch wrapped DEK via p2p
  |                              |                     |    from upload validator
  |                              |                     |    store in local validator store
  |                              |                     |                       |
  |-- PublishRelease tx (signed) ---------------------->|                       |
  |                              |                     |-- release.published -->|
  |                              |                     |                       |
```

### Properties

- **Phase separation.** The validator's `UploadComplete` and the client's `PublishRelease` are independent transactions. The file exists and is replicated before the client claims it. The client reviews the CIDs, verifies everything is correct, *then* signs.
- **Decoupled replication and access.** The replication set and DEK holder set are independent. Encrypted `.tdf` blobs can be replicated broadly for availability without exposing the DEK broadly. Validators in the replication set but not the DEK holder set store and seed ciphertext — they never see the raw key.
- **Lean chain state.** Only the uploader's wrapped DEK goes on-chain. Validator wrapped DEKs are distributed p2p and stored locally. Chain state per upload is O(1), not O(validators).
- **The client's signature is the entitlement.** The `PublishRelease` transaction is the cryptographic proof that this public key claims rights over this CID. The same key that can decrypt the file (because the DEK is wrapped against it) is the one asserting ownership.
- **CometBFT events drive replication.** Peer validators don't poll. They subscribe to `upload.complete` events, which contain the encrypted CID for BitTorrent rendezvous. Validators check whether they are in the replication set and/or DEK holder set to determine their role — replication set members fetch the `.tdf` blob, DEK holder set members additionally fetch their wrapped DEK from peers.

## Access / Consumption Flow

> **[Direct access diagram](diagrams/access-direct.mermaid)** · **[Forwarded access diagram](diagrams/access-forwarded.mermaid)**

### Direct (validator is in DEK holder set)

```
Device                    DEK Holder Validator         Chain
  |                          |                         |
  |-- access request:        |                         |
  |   { cid, device_x25519   |                         |
  |     _pubkey }             |                         |
  |   signed by Ed25519 ---->|                         |
  |                          |-- read policy for CID ->|
  |                          |<-- policy + conditions --|
  |                          |                         |
  |              verify Ed25519 signature              |
  |              verify consumer meets policy          |
  |                          |                         |
  |                          |-- GrantAccess tx ------>|
  |                          |   (audit: Ed25519 key   |
  |                          |    + CID + timestamp)   |
  |                          |                         |
  |<-- wrapped DEK           |                         |
  |   (wrapped to device's   |                         |
  |    X25519 pubkey)         |                         |
  |                          |                         |
  |  fetch .tdf via BitTorrent                         |
  |  unwrap DEK with device X25519 private key         |
  |  decrypt and play                                  |
```

### Forwarded (validator is in replication set but not DEK holder set)

```
Device                Replication Validator     DEK Holder Validator     Chain
  |                          |                         |                   |
  |-- access request ------->|                         |                   |
  |   (signed, with device   |                         |                   |
  |    X25519 pubkey)        |                         |                   |
  |                          |-- check DEK holder set ->|                   |
  |                          |   (not a DEK holder)    |                   |
  |                          |                         |                   |
  |<-- redirect to DEK       |                         |                   |
  |    holder validator -----|                         |                   |
  |                          |                         |                   |
  |-- access request -------------------------------->|                   |
  |                          |                         |-- read policy --->|
  |                          |                         |<-- policy --------|
  |                          |                         |                   |
  |                          |              verify signature + policy      |
  |                          |                         |                   |
  |                          |                         |-- GrantAccess --->|
  |                          |                         |   (audit only)    |
  |                          |                         |                   |
  |<-- wrapped DEK (to device X25519) ----------------|                   |
  |                          |                         |                   |
  |  fetch .tdf via BitTorrent                         |                   |
  |  (served by replication set — any seeding validator)|                   |
  |  unwrap DEK with device X25519 private key         |                   |
  |  decrypt and play                                  |                   |
```

The device decrypts locally. The validator never handles unencrypted content during consumption — it just brokers access by checking policy and wrapping the DEK to the device's key.

## Library Download

> **[Library download flow diagram](diagrams/library-download.mermaid)**

The model is iTunes, not Spotify — users download their library and own it locally. The download and decryption are separate steps: get the encrypted `.tdf` file first (safe to transfer openly), then get a wrapped DEK to unlock it.

### Clients

**Desktop (Tauri)** — the primary client. [Tauri](https://tauri.app/) v2, Rust backend, webview frontend (React/Svelte/etc.). The Rust core handles:

- BitTorrent: `librqbit` (pure Rust BT client) leeches `.tdf` blobs and PNGs directly from validators and good samaritans.
- Crypto: `ed25519-dalek`, `x25519-dalek` for signing and ECDH. `aes-gcm` for DEK-based decryption.
- Protobuf: `prost` generates Rust types from the same `.proto` files the validators use.
- Filesystem: full access. Library stored on disk wherever the user wants.

The Rust client talking to Go validators over protobuf proves the protocol is language-agnostic — the `.proto` files are the contract, not the implementation language.

**Browser** — the lightweight alternative. No install, no native code. Browsers can't speak native BitTorrent, so validators serve encrypted `.tdf` files over HTTP via the ConnectRPC/GraphQL API — validators are just HTTP file servers for ciphertext. For storage, the browser uses the **Origin Private File System (OPFS)** — persistent, sandboxed, available in all modern browsers. Files survive page reloads but live inside the browser sandbox.

WebTorrent (BitTorrent over WebRTC) can be layered on later as a P2P optimization. Validators would need to run WebTorrent-compatible seeding alongside native BitTorrent. Nice-to-have, not a requirement.

### How the `.tdf` file gets to the device

| Client | Transport | Storage |
|--------|-----------|---------|
| Tauri desktop | BitTorrent (native TCP/UDP) | User's filesystem |
| Browser | HTTP from validator API | OPFS (sandboxed) |

### Download flow

```
Device                    Validator / Peers
  |                          |
  |  1. query "what can I    |
  |     play?" via GraphQL   |
  |     (signed by Ed25519)  |
  |                          |
  |<-- list of CIDs + metadata
  |     + image CIDs (cover art)
  |                          |
  |  2. for each CID:        |
  |     download .tdf        |
  |     - browser: HTTP GET  |
  |     - native: BitTorrent |
  |     download PNGs        |
  |     - both: HTTP GET     |
  |     store locally        |
  |     (OPFS or filesystem) |
  |                          |
  |  3. to play a track:     |
  |     send access request  |
  |     (signed, with device |
  |      X25519 pubkey)      |
  |                          |
  |<-- wrapped DEK            |
  |                          |
  |  4. unwrap DEK locally   |
  |     decrypt .tdf → FLAC  |
  |     play                 |
```

Steps 2 and 3 can happen in either order. You can download your entire library in the background and request DEKs only when you're ready to play — or request access first and download on demand. The `.tdf` files are ciphertext, safe to have sitting on disk without a DEK.

### Offline playback

Once a device has both the `.tdf` file and a valid (non-expired) wrapped DEK, it can play offline. No network needed. For permanently granted content (purchased), the wrapped DEK never expires — true offline ownership. For subscription content with expiring DEKs, the device plays offline until the DEK expires, then needs to come online to refresh.

## Policy Plane

DDEX defines metadata formats but explicitly does not define access control — it assumes that happens at the infrastructure level (CDN buckets, web service auth, etc.). In a decentralized system, there is no infrastructure owner to enforce access. The policy plane fills this gap.

The policy plane has two layers: Casbin for structured IAM, and Goja for programmable logic that goes beyond what a policy engine can express.

### Casbin (structured IAM)

[Casbin](https://casbin.org/) is embedded as the primary policy engine. It provides RBAC, ABAC, ACL, and group-based access control with a mature, battle-tested evaluation engine. This is the right tool for IAM-shaped problems: defining a role, assigning it to a group, and having it cascade correctly. Anyone who has worked with AWS IAM, Kubernetes RBAC, or similar systems will recognize the model.

Casbin's default storage is file-based (`.conf` + `.csv`), which isn't suitable for on-chain state. This is solved by writing a custom Casbin adapter backed by CometBFT state. Policies and models are stored as structured on-chain data (indexable, queryable, composable), but evaluation uses Casbin's engine. We get both: clean chain state and real policy evaluation.

#### What Casbin handles

| Capability         | Example                                                                 |
|--------------------|-------------------------------------------------------------------------|
| Role assignment    | "Public key X has the `admin` role for catalog Y."                      |
| Group membership   | "Key X belongs to group `label-abc-staff`."                             |
| Role inheritance   | "`admin` inherits all permissions of `distributor` and `consumer`."     |
| Domain isolation   | "Label A's policies don't affect Label B's content."                    |
| Policy effects     | Allow/deny with first-match or all-match semantics.                     |
| ABAC attributes    | "Grant access if `request.territory` is in `policy.allowed_territories`." |

#### On-chain Casbin model

Each content owner (or their admin) can define a Casbin model and policies scoped to their content. A simple independent artist might use a default model:

```
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

A label with territorial distribution might use a richer model with domain-scoped roles and ABAC attributes. The model itself is stored on-chain and referenced by the content's access policy.

#### Policy types

| Policy type    | Backed by   | Example                                              |
|----------------|-------------|------------------------------------------------------|
| `public`       | (built-in)  | Anyone can access. DEK published in the clear.       |
| `direct_grant` | Casbin ACL  | Owner explicitly grants access to specific keys.     |
| `role_based`   | Casbin RBAC | Any entity with a `consumer` role can access.        |
| `conditional`  | Casbin ABAC | Access granted if request attributes match predicates (territory, time window, subscription tier). |
| `programmable` | Goja        | Custom JS logic for cases Casbin can't express.      |

The first four types cover the vast majority of use cases without any code execution — Casbin evaluates them as structured policy lookups.

### Goja (programmable policies)

When Casbin's declarative model can't express the logic needed, owners can deploy JavaScript policy scripts executed by the Goja runtime. This is the escape hatch, not the default path. It handles edge cases like:

- Complex multi-party licensing agreements with custom royalty splits
- Dynamic pricing based on on-chain state
- Cross-content bundle access (album logic)
- Custom attestation verification
- Conditional access based on external proofs submitted on-chain

A `SetAccessPolicy` transaction with type `programmable` references the CID of a deployed policy script. When a validator evaluates access, it loads the script and executes it in Goja.

Casbin can also delegate to Goja for custom matcher functions within an otherwise declarative model — a Casbin policy can reference a Goja function for a specific predicate without moving the entire evaluation into JavaScript.

## Goja Runtime

[Goja](https://github.com/dop251/goja) is an ECMAScript 5.1+ engine in pure Go. It embeds directly into the validator binary — no CGo, no external dependencies, runs on any platform Go supports. JavaScript is chosen over Rust/WASM because policy authors and proof authors are more likely to be application developers and label engineers than systems programmers. The barrier to entry matters.

Goja serves two roles in the system:

1. **Programmable policy evaluation** — access control logic that Casbin can't express.
2. **Proof generation** — computing attestations over on-chain activity data.

Both use the same runtime, the same determinism constraints, the same gas metering, and the same host API. Scripts are deployed once via `DeployScript` and referenced by either access policies or proof definitions.

### Determinism requirements

For consensus, every validator must produce the same output from the same inputs. The Goja runtime is constrained:

- **No `Date.now()`** — block timestamp is injected via `mojave.blockTime()`.
- **No `Math.random()`** — if randomness is needed, a deterministic PRNG seeded by block hash is injected.
- **No async / `setTimeout` / `setInterval`** — synchronous execution only.
- **No external I/O** — all state access through injected host functions.
- **No `eval()` or dynamic code loading** — the script is the script.

### Gas metering

Goja's `Interrupt()` mechanism is used for resource limiting. Each policy evaluation is allocated a gas budget (set as a chain parameter). An instruction counter increments on each JS operation; when gas is exhausted, execution halts and the policy evaluation fails closed (access denied). This prevents DoS via expensive policy scripts.

### Host API

The runtime exposes a `mojave` object with read-only state accessors:

```javascript
// State
mojave.blockTime()                          // current block timestamp (uint64)
mojave.blockHeight()                        // current block height (uint64)
mojave.getEntitlement(pubkey, cid)          // entitlement for a key over a CID
mojave.getRole(pubkey, cid)                 // role type for a key over a CID
mojave.getPolicy(cid)                       // access policy for a CID
mojave.getRelease(cid)                      // DDEX ERN metadata for a CID

// Activity (for proof scripts)
mojave.getAccessGrants(cid, from, to)       // GrantAccess txs for a CID in a range
mojave.getAccessGrantCount(cid, from, to)   // count of GrantAccess txs (cheaper than loading all)

// Cryptography
mojave.verify(pubkey, signature, msg)       // verify a cryptographic signature
mojave.hash(data)                           // deterministic hash
```

Policy scripts export an `evaluate` function that receives a request object and returns a decision:

```javascript
function evaluate(request) {
  // request.pubkey       — consumer's public key
  // request.cid          — content CID being requested
  // request.territory    — ISO 3166-1 territory code (if known)
  // request.attributes   — arbitrary key-value pairs from the request

  var entitlement = mojave.getEntitlement(request.pubkey, request.cid);

  if (!entitlement) {
    return { allowed: false, reason: "no entitlement" };
  }

  if (entitlement.role === "distributor") {
    if (entitlement.scope.territories.indexOf(request.territory) === -1) {
      return { allowed: false, reason: "outside licensed territory" };
    }
    return { allowed: true };
  }

  if (entitlement.role === "consumer") {
    if (entitlement.expiresAt < mojave.blockTime()) {
      return { allowed: false, reason: "access expired" };
    }
    return { allowed: true };
  }

  return { allowed: false, reason: "no matching rule" };
}
```

### Policy evaluation flow

> **[Policy evaluation diagram](diagrams/policy-evaluation.mermaid)**

When a validator processes a `GrantAccess` request:

1. Load the access policy for the requested CID from chain state.
2. If the policy type is `public`: grant immediately, no evaluation needed.
3. If the policy type is `direct_grant`, `role_based`, or `conditional`: evaluate via Casbin. The validator's embedded Casbin engine loads the model and policies from chain state (via the custom adapter) and runs `Enforce(sub, obj, act, ...attrs)`. No Goja involved.
4. If the policy type is `programmable`: load the referenced JS script, instantiate a Goja runtime with the `mojave` host API, call `evaluate(request)` with a gas budget.
5. If a Casbin model uses a custom matcher function backed by Goja: Casbin evaluates the policy as normal but calls into Goja for that specific predicate.
6. If the result is allowed: the validator wraps the DEK for the consumer's public key and submits a `GrantAccess` transaction.
7. If the result is denied or gas is exhausted: access denied, no transaction.

All validators must agree on the result. Casbin evaluation is deterministic given the same model and policies from chain state. Goja evaluation is deterministic given the same inputs and runtime constraints. Since all validators read the same chain state, they converge.

## Proofs & Attestations

> **[Attestation flow diagram](diagrams/attestation.mermaid)**

Not every user needs proofs, but for those who do — artists tracking plays for royalties, labels reporting to distributors, distributors proving delivery to licensors — the system needs a way to produce verifiable, portable claims about on-chain activity. This is where Goja fits most naturally: users deploy proof scripts that define what to prove, and validators produce signed attestations by running those scripts against chain state.

### What's already provable

`GrantAccess` transactions are on-chain. Each one is a signed, timestamped record that a specific consumer was granted access to a specific CID by a specific validator. From these, you can derive:

- Total access grants for a CID (or set of CIDs)
- Unique consumers per CID
- Access grants over a time range
- Access grants by territory (if territory is captured in the request)
- Access grants by distributor (which validator issued them, under which policy)

This is the raw data. Proof scripts aggregate and shape it into meaningful attestations.

### Access grants vs. play counts

There's a distinction between "access was granted" (a `GrantAccess` transaction happened) and "the file was actually played" (the consumer decrypted and listened). Once a consumer has a wrapped DEK, they can decrypt and play as many times as they want without further on-chain activity.

`GrantAccess` counts reflect unique access grants — closer to "unique listeners" than "total plays." For actual play counts, consumers or client applications would need to voluntarily report plays back to the chain (a `ReportPlay` transaction). This is opt-in and trust-the-reporter by nature, but it's useful for analytics and could be required by certain access policies. The system supports both: access grant proofs are trustless (derived from on-chain transactions), play count proofs are best-effort (derived from voluntary reports).

### Proof scripts

A proof script is a Goja function deployed on-chain via `DeployScript` and registered to specific content via `RegisterProof`. It defines:

1. **What to query** — which chain state to aggregate.
2. **How to compute the result** — counting, filtering, grouping.
3. **What shape the attestation takes** — the structured output document.

The script being on-chain means it's transparent and auditable. Anyone who receives an attestation can read the proof script to understand exactly what was computed and how.

```javascript
function prove(params) {
  // params.cid         — content CID to prove activity for
  // params.from        — start of time range (block height or timestamp)
  // params.to          — end of time range

  var grants = mojave.getAccessGrants(params.cid, params.from, params.to);

  var uniqueConsumers = {};
  var totalGrants = 0;
  var byTerritory = {};

  for (var i = 0; i < grants.length; i++) {
    totalGrants++;
    uniqueConsumers[grants[i].consumer] = true;

    var territory = grants[i].territory || "unknown";
    byTerritory[territory] = (byTerritory[territory] || 0) + 1;
  }

  return {
    cid: params.cid,
    from: params.from,
    to: params.to,
    totalGrants: totalGrants,
    uniqueConsumers: Object.keys(uniqueConsumers).length,
    byTerritory: byTerritory
  };
}
```

### Attestation flow

```
Client / Rights Holder         Validator                    Chain
  |                               |                           |
  |-- "run proof X for CID Y" -->|                           |
  |    (params: time range, etc.) |                           |
  |                               |-- load proof script X --->|
  |                               |<-- script source ---------|
  |                               |                           |
  |                               |-- query GrantAccess txs ->|
  |                               |<-- activity data ---------|
  |                               |                           |
  |                    execute proof script in Goja            |
  |                    sign attestation with validator key     |
  |                               |                           |
  |<-- signed attestation --------|                           |
  |                               |                           |
  |  (optionally)                 |                           |
  |                               |-- SubmitAttestation tx -->|
  |                               |                           |
```

The attestation document contains:

- The proof script CID (so the verifier knows what was computed)
- The block height at which the proof was evaluated (so the verifier can confirm the chain state)
- The computed result
- The validator's signature

### Off-chain portability

The signed attestation is a self-contained, verifiable document. Anyone who trusts the Mojave validator set can verify it:

1. Check the validator's signature against the known validator set.
2. Read the proof script from the chain (by CID) to understand what was proved.
3. Optionally, re-run the proof script against the same block height to independently verify the result.

Use cases for portable attestations:

- **Royalty payments**: "This track had 50,000 access grants in Q1" — present to a payment processor or label.
- **Distribution reporting**: "Here's the geographic breakdown of access for this catalog" — present to a licensor.
- **Grant applications**: "My independently released music had N unique listeners" — present as evidence.
- **Cross-chain interop**: present the attestation to a smart contract on another chain to trigger a payment, mint a token, or update state.

For stronger guarantees, a client can request the same proof from multiple validators and present a quorum of signatures. Since the proof script is deterministic and all validators read the same chain state, the results will match.

## Networking

> **[Networking layers diagram](diagrams/networking.mermaid)**

Three communication layers, each used where it fits:

| Layer | What it carries | Between whom | Why this tool |
|-------|----------------|-------------|---------------|
| CometBFT reactors | Small p2p messages — wrapped DEK distribution, DEK holder set sync, coordination | Validator ↔ validator | Peers are already connected and authenticated. Peer discovery, connection management, and message routing are handled by CometBFT. Custom reactors plug directly into this infrastructure. No separate p2p service needed. |
| BitTorrent | Large file transfer — encrypted `.tdf` blobs, PNGs | Validator ↔ validator (and consumers) | CometBFT's p2p is designed for small messages (txs, blocks, consensus). Bulk data needs a protocol built for it. Rendezvous on content CID. |
| ConnectRPC / GraphQL | Client-facing API — queries, transaction submission, file serving | Client ↔ validator/RPC | External API for UIs, SDKs, CLI tools. Not used for validator-to-validator communication. |

## API Layer

### Protobuf (transaction and p2p format)

All on-chain transactions and validator-to-validator p2p messages use Protocol Buffers as the wire format. Protobuf gives strong typing, backward-compatible schema evolution, and code generation for every major language. Transaction messages (`UploadComplete`, `PublishRelease`, `GrantAccess`, etc.) are defined as protobuf messages, signed by the submitter, and submitted to CometBFT via ABCI. Reactor messages (DEK distribution, sync requests) are also protobuf.

Protobuf is the internal language of the system — validators speak it, chain state is encoded in it, p2p reactor messages use it, and transactions are encoded in it. But protobuf is not what UIs should have to work with directly.

### Type generation pipeline

> **[Type generation pipeline diagram](diagrams/type-generation.mermaid)**

DDEX types are massive and live in a separate repo (`ddex-proto`). To avoid maintaining parallel type definitions across proto and GraphQL, the generation pipeline runs in two places:

**`ddex-proto` repo:**

1. `.proto` files define the DDEX ERN type hierarchy (source of truth).
2. `protoc` generates Go structs.
3. `protoc-gen-graphql` generates `.graphql` schema files from the same `.proto` definitions.
4. Both the Go types and `.graphql` schemas are published as part of the Go module.

**`mojave` repo:**

1. Imports Go types and generated `.graphql` schemas from `ddex-proto`.
2. Defines its own `.proto` files for Mojave-specific types (uploads, entitlements, policies, proofs, access grants) — these are small and manageable.
3. `protoc` generates Go structs + ConnectRPC services for Mojave types.
4. Hand-writes `.graphql` schemas for Mojave-specific types only.
5. `gqlgen` autobinds against both `ddex-proto` Go types and Mojave Go types. No duplicate type definitions.

The result: DDEX GraphQL types are auto-generated and stay in sync with proto automatically. Mojave-specific GraphQL types are hand-written but small. Nobody writes GraphQL schemas for the DDEX ERN hierarchy by hand.

### ConnectRPC + GraphQL (human API)

Validators and RPC nodes expose both APIs on top of the protobuf internals. The goal: a frontend developer with a wallet adapter (Phantom Connect, WalletConnect, etc.) should be able to build a simple iTunes-like library UI without running an indexer or understanding protobuf schemas.

**ConnectRPC** — gRPC-compatible but works over HTTP/1.1 and JSON out of the box. A browser can call it directly without a proxy. Generated from the same `.proto` files as the transaction format, so the API stays in sync with the chain schema automatically. Good for programmatic access: SDKs, CLI tools, validator-to-validator communication, transaction submission.

**GraphQL (gqlgen)** — flexible query language backed by [gqlgen](https://gqlgen.com/). Lets clients ask for exactly what they need in a single request. Good for UI-driven queries: "get all releases I own with their cover art, artist name, and access status." GraphiQL ships as the built-in API explorer — any developer can open a validator's GraphiQL endpoint and interactively browse the schema, run queries, and understand what's available without reading docs.

GraphQL queries are rich and typed. Mutations are not — there's essentially a single mutation: `sendTx(signedTransaction: bytes!): TxResult`. All writes go through CometBFT as signed protobuf transactions. The GraphQL layer doesn't try to model each transaction type as a separate mutation — that would be a leaky abstraction over what's fundamentally a blockchain write path. The client constructs a protobuf transaction, signs it with their wallet, and submits it. GraphQL is the read layer; protobuf + CometBFT is the write layer.

### What the API exposes (no indexing required)

The API is a read layer over chain state and validator local storage. Because all state is keyed by CID or public key, common queries are direct lookups — no indexing needed:

| Query | Backing data | Notes |
|-------|-------------|-------|
| "What do I own?" | Entitlements keyed by public key | Sign in with wallet, get all CIDs where your key has an `owner` entitlement. |
| "What can I play?" | GrantAccess audit records keyed by public key | All CIDs where your key has been granted access. The device requests a fresh wrapped DEK at play time. |
| "Get release metadata" | Release state keyed by CID | DDEX ERN, FLAC CID, image CIDs, status. |
| "Get cover art" | Image CID → PNG | Served directly by the validator. Unencrypted, no access gate. |
| "Get my access history" | GrantAccess txs keyed by public key | History of access grants for your key — when, what, which validator. |
| "Get policies for my content" | Access policies + Casbin state keyed by CID | For owners/admins managing their catalog. |

These are all O(1) or O(n) lookups against chain state where n is the number of items the user owns or has access to. No full-chain scans, no inverted indexes.

### What does require indexing

Some queries are inherently search-shaped and can't be answered by direct state lookups:

- Full-text search across release metadata ("find tracks by artist name")
- Discovery/browse ("new releases this week," "popular in your region")
- Analytics dashboards ("play counts across my catalog over time")

These are out of scope for the base validator API. They're better served by external indexers that subscribe to CometBFT events, build secondary indexes, and expose their own search APIs. The chain doesn't need to support them — it just needs to emit the events that make them possible.

### Wallet authentication

Clients authenticate by signing a challenge with their Ed25519 key. The API verifies the signature against the public key, then scopes all responses to that key's entitlements, grants, and policies. No sessions, no cookies, no passwords — just a signature. Compatible with any wallet that supports Ed25519 message signing (Phantom, etc.).

## Economics

> See [economics.md](economics.md) for the full token model, fee structure, validator rewards, and bootstrapping strategy.

Mojave has a native token (MOJ, base unit: grains, 1 MOJ = 10^9 grains) used for gas fees, storage fees, content purchases, and validator rewards. MOJ is the only on-chain currency; there is no attestation-based or multi-currency payment path in the protocol for now. Storage is the expensive operation — proportional to file size and replication factor. Gas for normal user activity (browsing, purchasing, policy changes) is effectively free. Content purchases are direct transfers in MOJ from consumer to content owner — the protocol takes no cut at the base layer. USD is a display and UX concern in clients (price feeds, “price in USD” with conversion to MOJ at purchase); liquidity and exchange (MOJ ↔ USD) are handled off-chain (faucet, then exchanges if they list MOJ).

## Design Principles

**Composability.** All state is on-chain and readable. DDEX ERNs, access policies, entitlements, CIDs — any application can index this and build on top. Music players, marketplaces, recommendation engines, royalty trackers — all just read chain state. No platform gatekeepers.

**Portability.** DDEX is an existing music industry standard. If a rights holder wants to leave, their metadata is in a format that Warner, Spotify, or any distributor already understands.

**Verifiability.** Every claim is signed. "I own this" — client signature on `PublishRelease`. "I processed this correctly" — validator signature on `UploadComplete`. "This access was authorized" — validator signature on `GrantAccess` + the on-chain policy that justified it. Full audit trail.

**Encryption is structural, not optional.** Files are *never* stored unencrypted on the network (only transiently during processing on the upload validator). Access is always mediated through policy-checked DEK wrapping. You can't accidentally make something public — you have to explicitly set a policy.

**CometBFT's validator set is the trust boundary.** The bounded, known validator set means you can pre-wrap DEKs for all validators at upload time. New validators joining the set can be handled by any existing validator creating wrapped DEKs for them (authorized by the validator set change event). The chain replaces the centralized KAS.

**Progressive complexity.** The policy plane is designed so that simple things are simple and complex things are possible. An independent artist setting a track to `public` is a single declarative transaction. A label managing territorial distribution rights across multiple partners can use role delegation and conditional policies. A complex multi-party licensing agreement that neither can express gets a custom JS policy script. Each level builds on the one below it.

## Governance

> See [governance.md](governance.md) for the full governance model — validator elections, oracle elections, copyright takedowns, and jurisdictional compliance.

Validators are admitted through social election (staking + community vote), not just staking alone. Oracles — elected per-network — handle copyright takedowns by submitting on-chain requests that trigger DEK removal. Jurisdictional compliance is a local validator decision, not a global consensus action. See the governance doc for details.

## Trust Assumptions

**Upload validator and unencrypted audio.** The upload validator temporarily holds unencrypted audio during transcoding. The validator is trusted not to leak the raw file during that window. In CometBFT, validators are bonded/staked and have economic skin in the game, which mitigates this. This is a conscious tradeoff, not an oversight.

**Policy script execution.** Policy scripts are user-deployed code running inside the validator. The Goja sandbox, gas metering, and restricted host API mitigate the risk, but a malicious or buggy script could still consume its full gas budget on every evaluation. Chain governance should set sensible gas limits and potentially require script audits or staking for deployment.

**Departing validators and DEK retention.** When a validator leaves the set, they retain any raw DEKs they previously unwrapped. This cannot be cryptographically revoked — once key material is known, it's known. The system mitigates this through layered mechanisms:

1. **Identity revocation.** The departing validator loses its active validator identity. It can no longer participate in `GrantAccess` or any chain operations. Even though it holds the DEK locally, it has no way to use it within the protocol.
2. **Unbonding period.** Validators must go through an unbonding period before their stake is returned (standard CometBFT/PoS pattern). During this window, malicious behavior (e.g. leaking decrypted content) can be detected and their bond slashed.
3. **Opt-in DEK rotation.** Content owners can request DEK rotation via `RotateDEK`: the file is re-encrypted with a fresh DEK, the new DEK is distributed to the current DEK holder set via p2p, and the new `.tdf` blob is re-seeded. This cryptographically revokes the ex-validator's access to that content.

The blast radius of a departing validator is directly controlled by the uploader's DEK holder set configuration. Content with a narrow DEK holder set (e.g. 5/60 validators) is minimally exposed — only 5 validators ever had the DEK, DEK rotation is cheap (re-wrap for 4 remaining + 1 replacement), and the majority of validators (the replication-only set) never had access to the raw key in the first place. They only store and seed ciphertext. Content with a wide DEK holder set (60/60) trades a larger blast radius for maximum access availability — the uploader accepts this tradeoff at upload time.

**Uploader recovery is a layer 2 concern.** The uploader's on-chain wrapped DEK is the last-resort recovery mechanism, but it's only needed if every validator in the DEK holder set goes down simultaneously — an extremely unlikely scenario with a well-chosen set. Even in that case, the encrypted `.tdf` blob is still in the replication set; the content isn't lost, just the ability to issue new wrapped DEKs. How the recovery key is managed (self-custody, custodial service, social recovery) is left to the uploader and can be solved at a layer above the protocol.
