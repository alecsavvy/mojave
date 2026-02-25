# Storage

Every validator runs two PebbleDB instances. They serve completely different purposes, have different durability guarantees, and never share data.

| | Chain Store | Local Store |
|---|---|---|
| Engine | PebbleDB (via CometBFT 1.x ABCI++) | PebbleDB (validator-managed) |
| Scope | Consensus — identical across all validators | Instance — unique to this validator |
| Durability | Replicated by consensus. Recoverable from any peer via state sync. | Ephemeral or locally durable. Lost if the node is destroyed. |
| Contents | DDEX metadata, entitlements, policies, uploads, uploader wrapped DEKs | Validator wrapped DEKs, processing scratch, local indexes |
| Who writes | ABCI++ state transitions (deterministic, consensus-gated; FinalizeBlock) | Validator application code (local, non-deterministic) |

## Chain Store

The chain store is CometBFT 1.x ABCI++ application state. Every validator holds an identical copy. Writes happen only through committed transactions (applied in FinalizeBlock) — a write that doesn't pass consensus doesn't exist. This is the source of truth for the entire network.

PebbleDB is the backing engine (CometBFT's default since the IAVL → PebbleDB migration). The ABCI++ application reads and writes through CometBFT's state management; it never opens the database directly.

### Key spaces

All keys are prefixed by domain to avoid collisions. Values are protobuf-encoded.

#### Uploads

**Prefix:** `uploads/`

**Key:** `uploads/{flac_cid}`

| Field | Type | Description |
|-------|------|-------------|
| `flac_cid` | bytes | Content hash of unencrypted transcoded FLAC — canonical identifier |
| `encrypted_cid` | bytes | Content hash of the encrypted `.flac.tdf` blob |
| `image_cids` | repeated bytes | Content hashes of associated artwork (PNG) |
| `uploader_pubkey` | bytes | Ed25519 public key of the uploader |
| `status` | enum | `PROCESSING_COMPLETE` · `PUBLISHED` |
| `timestamp` | uint64 | Block timestamp at creation |
| `replication_set` | repeated bytes | Validator pubkeys that store and seed the file |
| `dek_holder_set` | repeated bytes | Validator pubkeys that hold wrapped DEKs |

#### Releases

**Prefix:** `releases/`

**Key:** `releases/{flac_cid}`

| Field | Type | Description |
|-------|------|-------------|
| `flac_cid` | bytes | References the upload |
| `ddex_ern` | bytes | Full DDEX ERN metadata (protobuf-encoded, types from `ddex-proto`) |
| `signer_pubkey` | bytes | Ed25519 key that signed the `PublishRelease` tx — this is the entitlement |
| `signature` | bytes | The signature itself |
| `timestamp` | uint64 | Block timestamp at publication |

#### Entitlements

**Prefix:** `entitlements/`

**Key:** `entitlements/{cid}/{pubkey}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID this entitlement covers |
| `pubkey` | bytes | Ed25519 key that holds this entitlement |
| `role` | enum | `OWNER` · `ADMIN` · `DISTRIBUTOR` · `CONSUMER` |
| `grantor` | bytes | Pubkey that granted this entitlement |
| `scope` | message | Optional constraints — territories, time window, custom conditions |
| `created_at` | uint64 | Block timestamp |

**Secondary index:** `entitlements_by_key/{pubkey}/{cid}` → same value. Allows "what do I own?" queries by key without scanning all entitlements.

#### Access Policies

**Prefix:** `policies/`

**Key:** `policies/{cid}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID this policy covers |
| `type` | enum | `PUBLIC` · `DIRECT_GRANT` · `ROLE_BASED` · `CONDITIONAL` · `PROGRAMMABLE` · `TAKEDOWN_PENDING` · `TAKEN_DOWN` |
| `casbin_model_id` | string | References a Casbin model (for structured policy types) |
| `script_cid` | bytes | References a Goja script (for `PROGRAMMABLE` type) |
| `owner_pubkey` | bytes | Who set this policy |

#### Casbin Models

**Prefix:** `casbin/models/`

**Key:** `casbin/models/{model_id}`

| Field | Type | Description |
|-------|------|-------------|
| `model_id` | string | Unique identifier for this model |
| `owner_pubkey` | bytes | Content owner who defined this model |
| `perm_model` | string | The PERM model definition text (request_definition, policy_definition, etc.) |

#### Casbin Policy Rules

**Prefix:** `casbin/rules/`

**Key:** `casbin/rules/{model_id}/{rule_index}`

| Field | Type | Description |
|-------|------|-------------|
| `model_id` | string | Which model this rule belongs to |
| `ptype` | string | `p` (policy) or `g` (grouping/role assignment) |
| `v0..v5` | string | Rule fields — e.g. `(sub, obj, act)` or `(user, role, domain)` |

**Secondary index:** `casbin/rules_by_sub/{model_id}/{subject}` → rule keys. Allows loading all rules for a specific subject without scanning.

The custom Casbin adapter implements `persist.Adapter` (and optionally `persist.FilteredAdapter`) backed by these key spaces. Casbin's engine calls `LoadPolicy()` and `SavePolicy()` through the adapter, which reads/writes chain state. The adapter is read-only during policy evaluation (consensus has already committed the rules); writes happen only during `AddPolicy` / `RemovePolicy` transaction processing.

#### Wrapped DEKs (uploader only)

**Prefix:** `deks/`

**Key:** `deks/{cid}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID |
| `wrapped_dek` | bytes | DEK encrypted against the uploader's recovery X25519 key |
| `ephemeral_pubkey` | bytes | Ephemeral X25519 public key for ECDH unwrapping |
| `uploader_pubkey` | bytes | The uploader's Ed25519 identity key |

This is the only wrapped DEK that lives on-chain. It exists as a last-resort recovery mechanism.

#### Access Grants (audit log)

**Prefix:** `grants/`

**Key:** `grants/{cid}/{consumer_pubkey}/{block_height}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID |
| `consumer_pubkey` | bytes | Ed25519 key of the consumer granted access |
| `validator_pubkey` | bytes | Validator that issued the grant |
| `timestamp` | uint64 | Block timestamp |
| `block_height` | uint64 | Block height (for proof scripts that query by range) |

**Secondary index:** `grants_by_key/{consumer_pubkey}/{cid}/{block_height}` → grant key. Allows "what have I been granted?" queries.

#### Deployed Scripts

**Prefix:** `scripts/`

**Key:** `scripts/{script_cid}`

| Field | Type | Description |
|-------|------|-------------|
| `script_cid` | bytes | Content hash of the script source |
| `source` | bytes | JavaScript source code |
| `deployer_pubkey` | bytes | Who deployed it |
| `timestamp` | uint64 | Block timestamp |

#### Proof Definitions

**Prefix:** `proofs/`

**Key:** `proofs/{cid}/{proof_id}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID this proof covers |
| `proof_id` | string | Unique identifier for this proof definition |
| `script_cid` | bytes | References the deployed Goja script |
| `registrar_pubkey` | bytes | Who registered this proof |

#### Attestations

**Prefix:** `attestations/`

**Key:** `attestations/{proof_id}/{block_height}`

| Field | Type | Description |
|-------|------|-------------|
| `proof_id` | string | Which proof definition produced this |
| `block_height` | uint64 | Chain height at which the proof was evaluated |
| `result` | bytes | Protobuf-encoded proof result |
| `validator_pubkey` | bytes | Validator that produced and signed this |
| `signature` | bytes | Validator's Ed25519 signature over the result |

#### Accounts

**Prefix:** `accounts/`

**Key:** `accounts/{pubkey}`

| Field | Type | Description |
|-------|------|-------------|
| `pubkey` | bytes | Ed25519 public key |
| `nonce` | uint64 | Transaction sequence number (replay protection) |
| `created_at` | uint64 | First seen block timestamp |

There is no native token; payments use USDC attestations (see [economics.md](economics.md)). Accounts are identity + nonce only. Implicitly created on first transaction.

#### Validators (elected)

**Prefix:** `validators/`

**Key:** `validators/{pubkey}`

| Field | Type | Description |
|-------|------|-------------|
| `pubkey` | bytes | Validator's Ed25519 public key |
| `x25519_pubkey` | bytes | Validator's X25519 public key (for DEK wrapping) |
| `status` | enum | `CANDIDATE` · `ACTIVE` · `UNBONDING` · `REMOVED` |
| `stake` | uint64 | Optional bond or stake (TBD; no protocol token — see [economics.md](economics.md)) |
| `jurisdiction` | repeated string | ISO 3166-1 country codes |
| `identity` | string | Self-reported identity / organization name |
| `admitted_at` | uint64 | Block height when election passed |

#### Publisher groups (group registry)

**Prefix:** `groups/`

**Key:** `groups/{group_id}`

| Field | Type | Description |
|-------|------|-------------|
| `group_id` | bytes | Group identity (e.g. account or derived id) |
| `takedown_authority_pubkey` | bytes | Ed25519 pubkey(s) that can submit takedowns for this group's content |
| `status` | enum | `ACTIVE` · `SUSPENDED` · `REMOVED` |
| `stake` | uint64 | Optional bond/stake for group representation (TBD; no protocol token) |
| `admitted_at` | uint64 | Block height when `AdmitGroup` passed |

#### Takedowns

**Prefix:** `takedowns/`

**Key:** `takedowns/{cid}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID subject to takedown |
| `status` | enum | `PENDING` · `CONFIRMED` · `DISMISSED` |
| `group_id` | bytes | Publisher group that owns this content; only that group's takedown authority may submit |
| `takedown_authority_pubkey` | bytes | Pubkey that submitted the request (must match group's takedown authority) |
| `reason` | string | Takedown reason |
| `evidence_hash` | bytes | Hash of off-chain evidence |
| `requested_at` | uint64 | Block height of `TakedownRequest` |
| `resolved_at` | uint64 | Block height of `TakedownResolve` (if resolved) |
| `counter_notice` | bytes | Counter-notice content (if filed) |
| `counter_notice_by` | bytes | Pubkey that filed the counter-notice |

#### Content Flags (jurisdictional)

**Prefix:** `flags/`

**Key:** `flags/{cid}/{jurisdiction}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID |
| `jurisdiction` | string | ISO 3166-1 country code |
| `takedown_authority_pubkey` | bytes | Group takedown authority that set the flag |
| `reason` | string | Reason for jurisdictional restriction |
| `flagged_at` | uint64 | Block timestamp |

### State sync and recovery

A new validator joining the network doesn't replay every block from genesis. CometBFT's state sync snapshots the PebbleDB at periodic block heights. The new node downloads a recent snapshot, verifies it against the consensus, and starts applying blocks from that point forward. The chain store is fully recoverable from the network.

---

## Local Store

The local store is the validator's own PebbleDB instance, outside of consensus. Each validator's local store is different — it contains data specific to that validator's role in the network (what content it replicates, what DEKs it holds, what it's currently processing). No consensus on this data. If the node dies, the local store is rebuilt from peers.

### Key spaces

#### Validator Wrapped DEKs

**Prefix:** `local/deks/`

**Key:** `local/deks/{cid}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID |
| `wrapped_dek` | bytes | DEK encrypted against this validator's X25519 key |
| `ephemeral_pubkey` | bytes | Ephemeral X25519 public key for ECDH |

Populated via CometBFT reactors from the upload validator or existing DEK holders. The chain's DEK holder set configuration determines which CIDs this validator *should* have wrapped DEKs for. On startup, the validator compares its local DEK inventory against the chain state and requests missing wrapped DEKs from peers.

**Security:** This is the most sensitive data in the local store. The wrapped DEK is ciphertext (safe at rest — requires this validator's X25519 private key to unwrap), but it should still be stored in a restricted-access directory. The validator's X25519 private key itself lives in the system keychain or a secrets manager, never in PebbleDB.

#### Processing Scratch

**Prefix:** `local/processing/`

**Key:** `local/processing/{job_id}`

| Field | Type | Description |
|-------|------|-------------|
| `job_id` | string | UUID for this processing job |
| `status` | enum | `RECEIVING` · `TRANSCODING` · `ENCRYPTING` · `SEEDING` · `COMPLETE` · `FAILED` |
| `client_pubkey` | bytes | Uploader's Ed25519 key |
| `raw_audio_path` | string | Filesystem path to the incoming audio (temporary) |
| `flac_path` | string | Filesystem path to the transcoded FLAC (temporary) |
| `flac_cid` | bytes | Computed after transcoding |
| `encrypted_cid` | bytes | Computed after encryption |
| `image_paths` | repeated string | Filesystem paths to normalized PNGs (temporary) |
| `image_cids` | repeated bytes | Computed after normalization |
| `dek` | bytes | Raw DEK — **only exists during processing, wiped on completion** |
| `replication_set` | repeated bytes | Client-specified |
| `dek_holder_set` | repeated bytes | Client-specified |
| `created_at` | uint64 | Job start time |
| `error` | string | Error message if `FAILED` |

This is the upload pipeline's working state. The raw DEK exists here in plaintext only during the processing window — from DEK generation through encryption and wrapping. Once `UploadComplete` is submitted and all wrapped DEKs are distributed, the raw DEK is zeroed and the scratch entry transitions to `COMPLETE`.

**Security:** The `dek` field contains the raw DEK during processing. This is the one moment in the system where a raw DEK exists in memory and on disk. The processing scratch directory should have restricted permissions, and the validator is trusted not to leak it (see architecture.md Trust Assumptions).

On validator restart, incomplete processing jobs are either resumed (if the raw audio is still on disk) or failed and cleaned up. The client can retry the upload.

#### Sync State

**Prefix:** `local/sync/`

Tracks what this validator has fetched and what it still needs.

**Key:** `local/sync/content/{cid}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID (encrypted CID for .tdf, image CID for PNGs) |
| `status` | enum | `PENDING` · `DOWNLOADING` · `SEEDING` · `FAILED` |
| `content_path` | string | Filesystem path in the content store (see content.md) |
| `size_bytes` | uint64 | File size |
| `last_announce` | uint64 | Last BitTorrent announce timestamp |

**Key:** `local/sync/deks/{cid}`

| Field | Type | Description |
|-------|------|-------------|
| `cid` | bytes | Content CID for which a wrapped DEK is expected |
| `status` | enum | `PENDING` · `RECEIVED` · `FAILED` |
| `source_validator` | bytes | Pubkey of the validator that sent the wrapped DEK |

On startup and periodically during operation, the validator reconciles:

1. Check chain state for its membership in replication sets → ensure all expected content is in `local/sync/content/`.
2. Check chain state for its membership in DEK holder sets → ensure all expected wrapped DEKs are in `local/deks/`.
3. Anything missing transitions to `PENDING` and triggers a fetch from peers.

### Rebuilding the local store

The local store is not replicated by consensus, but it can be fully reconstructed:

1. **Wrapped DEKs**: request from existing DEK holders via CometBFT reactors. The chain state says which validators *should* hold wrapped DEKs — the new node asks them.
2. **Content files**: fetch via BitTorrent from the replication set. The encrypted CID is the rendezvous point.
3. **Processing scratch**: if there were in-flight uploads, they're lost. The client retries.
4. **Sync state**: rebuilt by reconciling against chain state on first startup.

The local store is a cache that can be warmed from the network. Losing it is inconvenient (re-download time), not catastrophic.
