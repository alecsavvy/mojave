# Content

This document covers the on-disk layout for content files — encrypted audio (`.flac.tdf`) and unencrypted images (`.png`). These files live outside of PebbleDB, managed by `gocloud.dev` as the storage abstraction, and served via BitTorrent and the API layer.

## File types

| Type | Extension | Encrypted | Format | Purpose |
|------|-----------|-----------|--------|---------|
| Audio | `.flac.tdf` | Yes (OpenTDF) | FLAC inside a `.tdf` container | Protected audio content |
| Image | `.png` | No | PNG | Cover art, artist photos — public |

All audio is normalized to FLAC before encryption. All images are normalized to PNG. Clients never need format negotiation.

The `.flac.tdf` extension communicates two things: the inner payload is FLAC, and the container is OpenTDF. A consumer who has the DEK unwraps the `.tdf` container and gets a `.flac` file.

## Directory layout

Content is stored under a configurable root directory. The default is `~/.mojave/content/`. The root is configurable via `--content-dir` or the `MOJAVE_CONTENT_DIR` environment variable, allowing operators to point it at a dedicated volume, a mount backed by `gocloud.dev` (S3, GCS, etc.), or any other storage backend.

```
{content_root}/
├── audio/
│   ├── {encrypted_cid_hex[0:2]}/
│   │   ├── {encrypted_cid_hex[2:4]}/
│   │   │   └── {encrypted_cid_hex}.flac.tdf
│   │   └── .../
│   └── .../
├── images/
│   ├── {image_cid_hex[0:2]}/
│   │   ├── {image_cid_hex[2:4]}/
│   │   │   └── {image_cid_hex}.png
│   │   └── .../
│   └── .../
└── tmp/
    └── processing/
        ├── {job_id}/
        │   ├── raw.upload        # incoming audio (original format)
        │   ├── transcoded.flac   # after FLAC transcoding
        │   ├── encrypted.flac.tdf # after encryption (moved to audio/ on completion)
        │   └── images/
        │       ├── 0.png         # normalized cover art
        │       ├── 1.png         # additional artwork
        │       └── ...
        └── .../
```

### Why two levels of sharding

CIDs are hex-encoded content hashes. A flat directory with millions of files kills filesystem performance on every OS. Two levels of sharding by the first four hex characters gives 256 × 256 = 65,536 buckets, keeping each directory small even at scale.

The shard prefix is derived from the CID itself — no lookup table, no indirection. Given a CID, the path is deterministic:

```
audio/{cid[0:2]}/{cid[2:4]}/{cid}.flac.tdf
images/{cid[0:2]}/{cid[2:4]}/{cid}.png
```

### `audio/`

Encrypted `.flac.tdf` files, keyed by their **encrypted CID** (the content hash of the `.tdf` blob itself, not the inner FLAC). This is the CID used for BitTorrent rendezvous — peers find each other by announcing the encrypted CID.

Files in `audio/` are ciphertext. They're safe to store on any backend, back up to cloud storage, replicate freely. Without the DEK they're opaque blobs.

### `images/`

Unencrypted `.png` files, keyed by their **image CID** (content hash of the PNG). Served directly over the API and BitTorrent without any access gate. Any UI can fetch cover art from any validator in the replication set.

### `tmp/processing/`

Scratch space for in-flight uploads. Each upload job gets its own directory named by `job_id` (UUID). Files move through stages:

1. `raw.upload` — the original audio file as received from the client (MP3, WAV, FLAC, whatever).
2. `transcoded.flac` — after FLAC transcoding. The FLAC CID is computed from this file.
3. `encrypted.flac.tdf` — after DEK generation and encryption. The encrypted CID is computed from this file.
4. `images/*.png` — normalized artwork. Image CIDs are computed from these files.

On completion:
- `encrypted.flac.tdf` is moved to `audio/{shard}/{shard}/{encrypted_cid}.flac.tdf`.
- Each `*.png` is moved to `images/{shard}/{shard}/{image_cid}.png`.
- The `raw.upload` and `transcoded.flac` are deleted. The validator does not retain unencrypted audio after processing.
- The `{job_id}/` directory is removed.

On failure or validator restart with incomplete jobs: the entire `{job_id}/` directory is cleaned up. The client retries.

## gocloud.dev integration

Validators use [gocloud.dev](https://gocloud.dev/) `blob.Bucket` as the storage abstraction. The directory layout above is logical — the actual backing store can be:

| Backend | URL scheme | Notes |
|---------|-----------|-------|
| Local filesystem | `file:///path/to/content` | Default. Good for single-machine validators. |
| AWS S3 | `s3://bucket-name?region=us-east-1` | For validators running in AWS. |
| Google Cloud Storage | `gs://bucket-name` | For validators running in GCP. |
| Azure Blob Storage | `azblob://container-name` | For validators running in Azure. |
| In-memory | `mem://` | Testing only. |

The validator's configuration specifies the `gocloud.dev` URL. All file operations (read, write, delete, list) go through the `blob.Bucket` interface. The directory structure and file naming are identical regardless of backend.

```go
bucket, err := blob.OpenBucket(ctx, "s3://mojave-content?region=us-east-1")
```

### Path construction

Given a CID, the content path is computed without any database lookup:

```go
func AudioPath(encryptedCID []byte) string {
    hex := hex.EncodeToString(encryptedCID)
    return fmt.Sprintf("audio/%s/%s/%s.flac.tdf", hex[:2], hex[2:4], hex)
}

func ImagePath(imageCID []byte) string {
    hex := hex.EncodeToString(imageCID)
    return fmt.Sprintf("images/%s/%s/%s.png", hex[:2], hex[2:4], hex)
}
```

No database lookup, no indirection. CID → path is a pure function.

## BitTorrent integration

At its core, Mojave's content layer is just a BitTorrent client. Every `.flac.tdf` and `.png` file lives in a standard BitTorrent swarm identified by its CID. Anyone who has the file can seed it. Anyone who knows the CID can leech it. The files are encrypted ciphertext (audio) or public images — there is no security risk in broad seeding.

This is how the system avoids the classic BitTorrent problem of dead seeds. Traditional BitTorrent relies on voluntary seeders who can disappear at any time. Mojave solves this by making a subset of seeders **obligated**: validators in a content's replication set are required to seed as a condition of their role in the network. They're staked, bonded, and slashable. Their seeding obligation is derived from on-chain state and verified by the reconciliation loop. They can't quietly stop seeding without it being detectable.

But validators aren't the only seeders — they're the guaranteed floor. The swarm is open:

| Participant | Seeding obligation | How they join |
|------------|-------------------|---------------|
| **Replication set validators** | Required. On-chain obligation, reconciliation-enforced. | Automatically via `upload.complete` event and reconciliation loop. |
| **Good samaritan validators** | Voluntary. No on-chain obligation. | Choose to seed content they're not assigned to — e.g. a validator that wants to improve availability for popular content. |
| **Good samaritan nodes** | Voluntary. Not even a validator. | Anyone can run a Mojave node (or just a plain BitTorrent client) and seed `.flac.tdf` files. No stake, no consensus participation, just bandwidth contribution. |
| **RPC nodes** | Voluntary. Serve content over HTTP. | Run the Mojave API stack without being a validator. Serve content from their local content store. Useful for CDN-like deployments. |
| **Desktop clients (Tauri)** | Voluntary. Leech and optionally re-seed. | After downloading a `.flac.tdf`, the client can continue seeding it to other peers via `librqbit`. The consumer becomes a seeder — classic BitTorrent behavior. |

The more popular a track is, the more seeders it has — consumers who downloaded it contribute bandwidth back. Validators guarantee a baseline; good samaritans and consumers amplify it. A track with 10,000 downloads has 10,000 potential seeders plus the validator floor. Dead seeds become a non-issue.

Good samaritan nodes and RPC operators don't need any special permission. The encrypted CID is public (it's on-chain in the upload record). The `.flac.tdf` file is ciphertext — seeding it doesn't grant access to the music. You need a DEK for that, and only DEK holder validators can issue one after checking the access policy. So broad seeding improves availability without compromising access control.

### Seeding

When any node has a content file (after processing, after fetching from peers, or after choosing to seed voluntarily):

1. Register the file with the local BitTorrent client.
2. Announce using the CID as the rendezvous identifier.
3. Serve pieces to any peer that requests them.

### Leeching

When any node needs a content file:

1. Look up the CID (from chain state, from an API query, or from a known reference).
2. Join the BitTorrent swarm using the CID as rendezvous.
3. Download from any seeding peer — validators, good samaritans, other consumers.
4. Write to the local content store at the deterministic path.
5. Optionally begin seeding.

### Client access

| Client | Audio (`.flac.tdf`) | Images (`.png`) |
|--------|-------------------|-----------------|
| Desktop (Tauri) | BitTorrent (native TCP/UDP via `librqbit`) | HTTP GET from validator API |
| Browser | HTTP GET from validator API | HTTP GET from validator API |

The API layer serves content files directly from the content store. For audio, this is ciphertext — no access check needed to serve the `.tdf` blob (the access check happens when the client requests a DEK). For images, it's plaintext — served without any gate.

```
GET /content/audio/{encrypted_cid}  →  audio/{shard}/{shard}/{cid}.flac.tdf
GET /content/images/{image_cid}     →  images/{shard}/{shard}/{cid}.png
```

The API routes map directly to content store paths. The validator reads from `gocloud.dev` and streams the response.

## Reconciliation — what should I be seeding?

A validator's seeding obligations come from chain state, not from what's on disk. The content store is a consequence of the chain store — the validator crawls PebbleDB to figure out what it should have, diffs against what it actually has, and fills the gaps.

### The reconciliation loop

Runs on startup and periodically (configurable interval, default 60 seconds):

```
1. Scan chain store: iterate uploads/ prefix in PebbleDB
2. For each upload:
   a. Deserialize the upload record
   b. Check if this validator's pubkey is in the replication_set
   c. If yes: this validator SHOULD have the .flac.tdf and all .png files
3. Build the "expected set" — a map of CIDs this validator is responsible for
4. Scan local store: iterate local/sync/content/ prefix in PebbleDB
5. Build the "have set" — CIDs with status SEEDING
6. Diff:
   - expected − have = MISSING → trigger fetch
   - have − expected = ORPHANED → trigger cleanup
```

### Determining seeding obligations from PebbleDB

The chain store's `uploads/` key space is the authoritative source. Each upload record contains the `replication_set` — a list of validator pubkeys. The validator checks its own pubkey against this list:

```go
func (v *Validator) shouldReplicate(upload *Upload) bool {
    for _, pk := range upload.ReplicationSet {
        if bytes.Equal(pk, v.pubkey) {
            return true
        }
    }
    return false
}
```

For each upload where `shouldReplicate` returns true, the validator needs:

| CID | File | Path |
|-----|------|------|
| `upload.encrypted_cid` | Encrypted audio | `audio/{shard}/{shard}/{encrypted_cid}.flac.tdf` |
| each `upload.image_cids[i]` | Cover art / artwork | `images/{shard}/{shard}/{image_cid}.png` |

The validator checks its local sync state (`local/sync/content/{cid}`) and the content store (`gocloud.dev` existence check) to determine if each file is present.

### Rendezvous — CID as the BitTorrent join key

The CID is both the content address and the BitTorrent rendezvous identifier. When a validator determines it's missing a file, it joins the swarm using the CID:

```go
func (v *Validator) fetchMissing(cid []byte, contentType ContentType) error {
    path := contentPath(cid, contentType)

    if exists, _ := v.bucket.Exists(ctx, path); exists {
        return nil
    }

    v.localStore.Put(syncContentKey(cid), &SyncEntry{
        CID:    cid,
        Status: DOWNLOADING,
    })

    infohash := deriveInfohash(cid)
    v.torrentClient.AddTorrentFromInfohash(infohash, path)

    return nil
}
```

The infohash is derived deterministically from the CID — every validator computes the same infohash for the same content, so they find each other in the DHT or via tracker without coordination. The `.flac.tdf` filename on disk matches the encrypted CID, so the path is recoverable from the CID alone.

### Handling the diff

**Missing files** (expected but not present):

1. Create a `local/sync/content/{cid}` entry with status `PENDING`.
2. Derive the infohash from the CID.
3. Add to the BitTorrent client — it joins the swarm, downloads from seeders, writes to the deterministic content path.
4. On completion: update sync entry to `SEEDING`, register with the BitTorrent client for ongoing seeding.

**Orphaned files** (present but no longer expected — validator was removed from the replication set):

1. Stop seeding via BitTorrent.
2. Delete the file from the content store.
3. Remove the `local/sync/content/{cid}` entry.

Orphan cleanup is safe because the content is ciphertext and the validator has no obligation to retain it. The file still exists on other validators in the replication set.

**Corrupted files** (present but CID doesn't match content hash):

1. The reconciliation loop can optionally verify file integrity by recomputing the content hash and comparing to the CID.
2. If mismatched: treat as missing — delete and re-fetch.
3. This is expensive (reads every file) so it runs on a slower schedule (e.g. daily) or on-demand.

### Startup vs. steady-state

**Startup (cold start or state sync):**

The validator has just synced chain state but has an empty content store. The reconciliation loop runs immediately and queues every expected CID for download. This is the heaviest load — potentially thousands of torrents starting simultaneously. The BitTorrent client should rate-limit concurrent downloads to avoid saturating the network or disk.

**Steady-state (event-driven):**

During normal operation, the validator doesn't need to crawl `uploads/` on every tick. New replication obligations arrive via CometBFT events:

- `upload.complete` — new content to replicate. The validator checks if it's in the replication set and fetches immediately.
- `replication_set.updated` — a content owner changed the replication set. The validator checks if it was added or removed.

The periodic reconciliation loop is a safety net — it catches anything missed by events (e.g. events lost during a restart, race conditions). The event-driven path handles the common case with low latency.

### DEK holder reconciliation

The same pattern applies to DEK holder obligations, but against the `dek_holder_set` field and the `local/deks/` key space in the local store. If this validator should hold a wrapped DEK for a CID but doesn't have one, it requests it from a peer DEK holder via CometBFT reactors (not BitTorrent — wrapped DEKs are small messages, not bulk data).

## Lifecycle

### Upload (upload validator)

```
1. Client sends raw audio + images
2. Write raw audio to tmp/processing/{job_id}/raw.upload
3. Transcode to tmp/processing/{job_id}/transcoded.flac
4. Compute FLAC CID
5. Normalize images to tmp/processing/{job_id}/images/*.png
6. Compute image CIDs
7. Generate DEK, encrypt to tmp/processing/{job_id}/encrypted.flac.tdf
8. Compute encrypted CID
9. Move encrypted.flac.tdf → audio/{shard}/{shard}/{encrypted_cid}.flac.tdf
10. Move *.png → images/{shard}/{shard}/{image_cid}.png
11. Delete tmp/processing/{job_id}/
12. Begin seeding via BitTorrent
13. Submit UploadComplete tx
```

### Replication (peer validator)

```
1. Receive upload.complete event from chain
2. Check if this validator is in the replication set
3. If yes: join BitTorrent swarm for encrypted CID
4. Download .flac.tdf → audio/{shard}/{shard}/{encrypted_cid}.flac.tdf
5. Join BitTorrent swarm for each image CID
6. Download .png → images/{shard}/{shard}/{image_cid}.png
7. Begin seeding both
```

### Deletion

Content deletion is governance-scoped — individual validators can remove files they're no longer required to store (e.g. removed from a replication set), but the protocol doesn't have a unilateral "delete from the network" operation. A content owner can update their replication set to remove validators, and removed validators clean up their local copies.

```
1. Validator detects it's no longer in the replication set for a CID
2. Stop seeding via BitTorrent
3. Delete audio/{shard}/{shard}/{encrypted_cid}.flac.tdf
4. Delete images/{shard}/{shard}/{image_cid}.png (for each associated image)
5. Remove local/sync/content/{cid} from local store
```

## Disk usage

Rough sizing for capacity planning:

| Content type | Typical size | Notes |
|-------------|-------------|-------|
| `.flac.tdf` (single track) | 25–50 MB | FLAC is lossless. TDF adds ~1 KB overhead. |
| `.png` (cover art) | 0.5–5 MB | Depends on resolution. No resizing is done — stored as-is after PNG normalization. |

A validator in the replication set for 10,000 tracks with cover art: ~250–500 GB audio + ~5–50 GB images.

The content directory is the largest disk consumer by far. The chain store and local PebbleDB are negligible by comparison (metadata, not media).
