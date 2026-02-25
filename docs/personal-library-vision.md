# Reorientation: Decentralized Personal Music Library

This document describes reorienting Mojave from **decentralized distribution for artists/labels** toward **your decentralized music library** — a single place where you own your rips, buy from artists, and (in the future) trade or resell digital copies. The protocol remains infrastructure; the framing shifts to the consumer.

## Vision shift

| Before (distribution-centric) | After (library-centric) |
|-------------------------------|--------------------------|
| "Protocol for artist-to-fan distribution" | "Your music library: rip it, buy it, own it" |
| Primary actor: artist/label uploading and selling | Primary actor: **you** — building and owning your collection |
| Consumer = buyer of catalog content | Consumer = **library owner** — rips + purchases + (future) resales/trades |
| Chain as distribution ledger | Chain as **library ledger** — what you have, where it came from, what you can do with it |

Artist-to-fan sales remain a first-class use case; they become one of **how your library grows**, not the only one.

## Three ways your library grows

### 1. Digitize your personal library

You should be able to **rip or import** music you already own and store it in Mojave:

- **Vinyl** — rip to FLAC (or upload a rip you made elsewhere), encrypt, store. Your copy; only your devices get the DEK.
- **CDs** — same: rip or import, encrypt, store.
- **Digital copies** — files you already have (purchased elsewhere, Bandcamp, etc.): import, normalize to FLAC, encrypt, store.

**Properties:**

- **Private by default.** Personal-library content is not in the public catalog. It is keyed by (owner, CID) or a dedicated namespace (e.g. `library/{owner_pubkey}/...`). No discovery by others, no takedown by publisher groups — it’s your private storage.
- **You pay storage.** You’re the only “rights holder”; you pay for replication when you want validator backup (payment model deferred; assumed USDC subscription by content amount). Replication set can be small (e.g. 1–3 validators) or a “personal vault” tier.
- **You hold the only DEK.** No DEK holder set of validators in the distribution sense. Either:
  - **Option A:** Client-side encryption — you generate the DEK and wrap it to your recovery key (and optionally to a validator for durability). Validators never see plaintext; they only store ciphertext and, if desired, a backup wrapped DEK for recovery.
  - **Option B:** Upload via a validator (like today) but the content is tagged as `personal_library`; the only entitlement is owner = you; no replication to a broad set, no DEK distribution to other validators — just your wrapped DEK on-chain and ciphertext replicated for durability.

Option A is stronger for privacy (validator never sees your rip); Option B reuses current upload pipeline with a different policy/namespace. Design choice: see open questions below.

- **Metadata.** You don't supply DDEX yourself. The main UX is to **associate your copy with the artist's catalog release** (use their DDEX for display; your file is your file — see [Provenance, association, and what you can sell](#provenance-association-and-what-you-can-sell)). Optional: minimal metadata only if you don't link to a release; no DDEX ERN required for library-only items. Link to a catalog release (e.g. “this rip matches catalog CID X”) for display/grouping without giving anyone else access.

### 2. Purchase from artists (and sellers)

Unchanged from current design:

- **Catalog content** — artist/label uploads, publishes release, sets price and policy.
- **You buy** — `GrantAccess` (with payment) gives you a consumer entitlement; validator issues wrapped DEK to your device; you download `.flac.tdf` and play.
- **You own it** — permanent entitlement, offline playback, no subscription.

From the library perspective, “purchased” is a **source** for a library item: same as “ripped” or “imported,” but the chain records that you obtained it via a sale (and who was paid).

### 3. Future: Digital trades and secondhand sales

Design space to reserve; not required for v1. **Only purchased entitlements** can ever be traded or resold; rips and imports are never sellable (see [Provenance, association, and what you can sell](#provenance-association-and-what-you-can-sell)). Resale can be **governed** (white-glove, like iTunes) rather than permissionless.

- **Trades** — two users swap entitlements (e.g. “my track A for your track B”). Needs:
  - Transfer of consumer entitlement, or a new transaction type (e.g. `ProposeTrade`, `AcceptTrade`) that atomically revokes one party’s access and grants the other.
  - Clear rules: only transferable entitlements (e.g. “download” not “subscription”), and possibly artist/label policy (e.g. “this release is resaleable” in DDEX or policy).
- **Secondhand sales** — you sell your copy to another user. Needs:
  - Same transfer mechanism.
  - Optional: royalty or fee to original rights holder (policy or Goja script); or first-sale style (no ongoing cut).
  - UX: “Sell this track” → list for MOJ → buyer pays you → entitlement transfers.

The existing **TransferEntitlement** (owner transfer) is for catalog ownership (e.g. label sells catalog). For consumer-to-consumer we need **transfer of consumer entitlement** or a dedicated “resale/trade” flow so the chain and clients can represent “I used to have it, now they do.”

## Provenance, association, and what you can sell

### No proof of physical origin

There is **no way to cryptographically ensure** that a file came from a legit hard copy. CDs can be deterministic (same pressing → same bits → same hash), so in theory a catalog release and a perfect CD rip could share a FLAC CID; that's coincidence, not proof of ownership. **Vinyl is never deterministic** — every rip is unique (pressing, wear, ADC). So we don't try to "verify" that your file came from physical media. Rips and imports are **your copy**; the chain records source = rip/import and who owns it, not where the bits came from.

### Associate your copy with a release (metadata, your file)

The useful model: **use release metadata** (title, artist, artwork) and **provide your own file**. You're not claiming your file *is* the catalog file — you're saying "this is my copy of *that* release." Metadata can come from the artist's published DDEX (when they're on-chain) or from **public sources** (MusicBrainz, Discogs, etc.) — no label connection required (see below). So:

- When the artist/label has published on-chain: you add a **personal library item** (your rip or import) linked by **catalog reference** (e.g. `catalog_release_cid`). Your item uses their DDEX and artwork for display; the actual audio is your file (your CID, your DEK).
- When they haven't: you link by **public release ID** (e.g. MusicBrainz MBID, Discogs ID) and the client or chain stores/refers to public metadata for display. Same UX — your file, their (public) metadata.
- In both cases you get the same "white glove" experience; the chain and policy know **source = rip/import**, so it's for personal use only.

No need to prove the rip came from a CD or vinyl; the link is "I'm associating my copy with this release," not "I'm proving I own the physical medium."

**Public metadata (no label connection required).** You can associate your copy with release metadata from **public sources** — MusicBrainz, Discogs, or other open catalogs — without any connection to record labels or to a catalog release on-chain. When an artist/label has published that release on Mojave, you can link to their DDEX and reuse their artwork/metadata; when they haven’t, you still get a “white glove” display by pulling public metadata (title, artist, cover art reference, etc.). So personal library metadata is not dependent on labels being on the chain; it’s “this is my copy of *that* release” where “that release” can be identified by an on-chain catalog reference *or* by a public identifier (e.g. MusicBrainz MBID, Discogs release ID). No license or relationship to the label is implied for the file itself — it’s your copy, your encryption, your storage; the metadata is for display only.

### Validators hosting your encrypted copy: backup, not distribution

Validators that store personal library content hold **ciphertext** that only the uploading user can decrypt (only their key can unwrap the DEK). So a validator hosting a “Green Day” album that is encrypted only to the user who uploaded it is **not distributing** the work: no one else can access it. Functionally, the validator is an **encrypted backup / cloud storage** provider for that user’s file — like the user storing an encrypted blob in S3 or a vault service. The network does not grant access to third parties; it holds opaque ciphertext and, on-chain, a wrapped DEK that only the owner’s key can unwrap. So it’s **technically not distribution** in the copyright sense; it’s durable, encrypted storage for the user’s own copy. That framing is explicit in the design: personal library is private, single-owner, and access-controlled so that only the owner ever receives a DEK. Governance and policy can treat it as such (e.g. validators are not “distributing” personal library content; they’re providing encrypted backup for the paying user).

**Contrast with raw BitTorrent (DMCA).** Raw BitTorrent and similar P2P file-sharing get DMCA takedowns because the service or users **distribute** the work — they make it available to third parties who can download and decrypt. The operator is facilitating access to copyrighted content by the public. Personal library on Mojave is designed to create a **different factual and legal posture**: validators hold ciphertext that **only the account-holding user** can decrypt; no third party can ever access the work. The validator does not "make available" the work to anyone else; it provides encrypted storage for that user's copy, analogous to a cloud backup provider holding client-side-encrypted data. So there is a **legal argument** that this is not "distribution" under copyright law and that validators are not liable as distributors — they are storage providers for user-owned, user-encrypted content. (This is design intent and a plausible legal distinction, not legal advice; outcomes depend on jurisdiction and how courts apply the facts.)

### Ripping doesn't allow you to sell

Only **purchased** entitlements (and only when marked transferable by policy/DDEX) can ever be resold or traded. Rips and imports:

- Are **never** sellable or tradeable. Source = rip | import is a hard constraint: no transfer of consumer entitlement for those items.
- Don't grant any distribution rights; they're private to your library.

So the chain distinguishes **source** (rip, import, purchase, and later trade/resale). Resale and trade flows only apply to items whose source is purchase (and whose policy allows transfer). Governance can tighten further (e.g. only certain deal types, only after a holding period).

### White-glove resale through governance

Like iTunes' curated selling experience, **resale can be governed** rather than permissionless:

- **What is sellable:** Only items you **purchased** on-chain (consumer entitlement from a catalog sale). Not rips, not imports, not "I have a file and linked it to a release."
- **Who says it's resaleable:** The catalog release's policy or DDEX (e.g. "consumer entitlement is transferable") so artists/labels opt in.
- **How it runs:** Governance can define the resale/trade flow (e.g. listing, escrow, royalty to rights holder, validator behavior). That keeps secondhand a first-class but controlled experience — white glove, not a free-for-all.

So: associate your rip with the artist's DDEX for a great library UX; your file stays yours and is never sellable. Only purchased, transferable entitlements can move in a governed resale/trade flow.

### Replication and who pays

**Catalog (artists selling):** Artists and labels pay validators for replication and distribution. They choose replication set size; payment for replication is deferred (assumed USDC subscription by content amount). Validators earn for storing and seeding catalog content. That’s the main “pay for replication” path.

**Personal library:** Your devices come first. You can **sync from your own devices** (e.g. Tauri desktop seeds, phone leeches) without paying validators. If you want backup or sync when your desktop is off, you **optionally subscribe** (assumed USDC subscription by content amount) so validators seed your library blobs. So: free path = device-to-device sync; paid path = validator-backed backup via subscription.

### Scalability: deduplication and storage

We don’t want a million slight variations of the same album on the network. Options:

- **Deduplicate when FLAC CID matches.** Before storing a new personal-library blob, check if that FLAC CID already exists (catalog or another user’s library). If it does: don’t store a new blob; create a library entry that references the existing blob and issue a wrapped DEK for this user. One blob, many owners. Same FLAC (e.g. same CD rip) ⇒ one stored ciphertext.
- **Reference catalog when possible.** If the user’s rip has the same FLAC CID as a catalog release, treat it as “entitlement to the catalog copy” — no new storage; user gets access to the existing catalog blob (and can be recorded as “source = rip, matched catalog” for UX).
- **Small replication for personal library.** When we do store a new library blob, use a small replication set (e.g. 1–3 validators) so we’re not multiplying cost.
- **Optional: convergent encryption.** Derive DEK deterministically from FLAC (same content ⇒ same ciphertext). One blob per unique FLAC; wrap that DEK to each owner. Tradeoff: “same file” is visible; for music that may be acceptable.

Vinyl rips rarely deduplicate (every rip is unique); CD rips and imports often can. Design so that “same FLAC ⇒ one blob” is the default when we can detect it.

### Device sync: your devices first, validators when you pay

The **Tauri desktop app** can seed your library (`.tdf` blobs) via BitTorrent. Your **phone** (or another device) can leech from your desktop when they’re on the same network — same encrypted CID, same swarm. Local peer discovery (e.g. multicast on LAN) makes it easy for the phone to find the laptop on the same Wi‑Fi. When the phone is on cellular or the desktop is off, it can’t reach your desktop; that’s when **paying validators to seed** your library gives you sync and backup. So: sync from your own devices when possible; pay validators for durability and for sync when your desktop isn’t available.

### Physical ownership and digital entitlement

Ideally: prove you own a CD or vinyl and get entitled to the artist’s digital copy. **You can’t prove physical ownership from the audio alone** — CD same-bits could be from a download; vinyl is never deterministic. So:

- **Redeem codes.** Artists/labels ship a one-time (or limited) code with physical media. You redeem it on-chain; the protocol grants you a consumer entitlement to the catalog release. That’s real proof of physical ownership: only someone with the physical product has the code. Protocol supports a “redeem code C for release R” flow.
- **Same FLAC as catalog (CD), honor-system.** If your CD rip has the same FLAC CID as the catalog release, the protocol or policy can grant you the same entitlement as a buyer (no payment, same DEK). Convenient for people who bought the CD; not secure proof (anyone with the same file could claim it). Artists can opt in as a convenience.
- **Vinyl.** No content-based proof. Would need redeem codes or other out-of-band verification (e.g. serial, NFC) if we want “own vinyl → get digital.”

### Optional: acoustic match (local AI)

A **local** (or optionally server-side) model can answer “is this rip the same track/release as that reference?” — acoustic fingerprinting or similarity. Use cases:

- **UX:** “Your rip likely matches Green Day – American Idiot; use this metadata and associate?” Improves the “associate your copy with a release” flow.
- **Deduplication (optional):** If we ever treat “acoustically same” as same content, we could avoid storing a new blob for vinyl rips that match a reference. That would require the same model run in a trusted context (e.g. validator or dedicated service); privacy tradeoff (they process your audio). Client-side only: good for UX hints; not trusted for on-chain dedup (user could game it).

So: local AI for “same release?” is **probabilistic**, not proof. Useful for association and possibly for server-side dedup if we accept the privacy tradeoff.

## What stays the same

- **Consensus, storage, encryption planes** — CometBFT, PebbleDB, OpenTDF, BitTorrent, gocloud.dev unchanged.
- **Keys** — Ed25519 identity, X25519 device-scoped; same DEK wrapping model for access.
- **Catalog flow** — upload → UploadComplete → PublishRelease → SetAccessPolicy → GrantAccess (with optional payment). No change for artist/label distribution.
- **Economics** — No native token. **USDC attestations** for user subscriptions (library size) and content purchases. Artists sign up as distributors; grant access on purchase attestation. Validators take a cut; artist sale volume + user subscription fees pay for hosting. Users and good samaritans also seed (BitTorrent).
- **Governance** — Validator elections, publisher groups, takedowns. Takedowns apply to **catalog** content only; personal library is out of scope (private, not discoverable).

## What we add (design)

### Content source / namespace

- **Catalog** — `uploads/`, `releases/`, entitlements, DEK holder set, replication set, policies, takedowns. Existing key spaces.
- **Personal library** — same or new key spaces with a clear distinction:
  - Option: new prefix `library/{owner_pubkey}/{flac_cid}` (or by encrypted CID) for “library item” state: owner, storage fee payer, source = `rip` | `import` | `purchase` (and if purchase, reference to catalog CID and seller).
  - Entitlements for library items: only `owner` (you); no distributor/consumer roles. Access = you request DEK; validator checks you’re the owner and issues wrapped DEK (same as today for catalog owner).

### Flows to document (and implement)

1. **Rip/import (personal library)**  
   - Client: upload raw audio (or pre-normalized FLAC) + optional metadata.  
   - Tag: `personal_library`, `owner_pubkey`, optional `catalog_cid` (for “matches this release” display).  
   - Validator: same pipeline (transcode → FLAC, normalize images, encrypt) **or** client-side encrypt and upload ciphertext only.  
   - Chain: record under library namespace; only owner’s wrapped DEK on-chain; replication set small; no DEK holder set (or single “vault” validator).  
   - Payment for replication deferred (assumed USDC subscription by content amount when owner wants validator backup).

2. **Purchase (unchanged)**  
   - Already defined: GrantAccess with payment → consumer entitlement → DEK to device.

3. **“My library” query**  
   - Today: “What do I own?” (owner entitlements) + “What can I play?” (consumer grants).  
   - After: “What’s in my library?” = union of: (a) catalog items I own (owner), (b) catalog items I can play (consumer), (c) **personal library items** (owner in library namespace).  
   - Each item has a **source**: rip, import, purchase, (future) trade, resale.

### Future: transfer of consumer entitlement

- **TransferConsumerEntitlement** or **Resale** / **Trade** transaction types.  
- Rules: only for entitlements that are marked transferable (e.g. by policy or DDEX); atomic: revoke from A, grant to B; optional payment (resale) or swap (trade).  
- Economics: optional royalty to original rights holder (policy/script); protocol fee or none — TBD.

## Open questions

1. **Client-side vs validator-side encryption for personal library**  
   Client-side: maximum privacy (validator never sees audio); requires client to do FLAC normalization and encryption (or trust a dedicated “personal vault” service). Validator-side: reuses current upload path; validator still sees plaintext during processing. Tradeoff: privacy vs implementation reuse.

2. **Replication for personal library**  
   Minimal replication (e.g. 1–3 validators) vs “personal vault” (dedicated validators or tier) vs same replication as catalog. Affects cost and durability.

3. **Metadata for rips**  
   Optional only, or require minimal (e.g. content hash + owner)? Linking to catalog release (for “this rip is the same as that release”) is optional but useful for UX.

4. **Secondhand / trade policy**  
   Who decides if a given catalog item is resaleable or tradeable? DDEX deal type, on-chain policy, or Goja script. Need a clear rule so validators can allow or deny transfer.

5. **Discovery**  
   Personal library is not discoverable by others. Catalog remains discoverable (indexers, search). No change to discovery for catalog.

## Suggested doc and UX changes

- **docs/README.md** — Reframe opening: “Mojave is your decentralized music library: digitize your collection (rips, imports), buy from artists, own your files. In the future: trade or resell. Same protocol for catalog distribution and personal storage.” Keep “what this enables” for both artists and **library owners** (rips, purchases, one place for everything).
- **docs/architecture.md** — Add section **Personal library** (or **Content sources**): catalog vs personal library; flows for rip/import; “my library” as union of owner catalog + consumer catalog + library namespace; reserve **Transfer consumer entitlement / resale / trade** as future.
- **PROTOCOL.md** — Add short subsection: “Library ownership” — how a client asks “what’s in my library” (catalog + library items), and how rip/import fits (tag, storage fee, private).
- **docs/storage.md** — When we add it: new key spaces or prefixes for `library/` and any “source” or “transferability” fields.
- **docs/economics.md** — Storage fees for personal library (payer = owner); no content purchase for your own rips; future: optional resale/trade fees or royalties.
- **docs/governance.md** — Clarify: takedowns apply only to catalog content; personal library is private and not subject to group takedown (unless we later define abuse policy for shared library content).

## Summary

Reorienting to **your decentralized music library** means:

1. **Consumer-first framing** — “your library” that grows by ripping/importing, buying, and (later) trading/reselling. One app = library + player (no “Bandcamp then VLC”).
2. **Personal library as a first-class source** — rips and imports stored privately; you pay storage only when you want validator backup. Sync from your own devices first (desktop seeds, phone leeches); optionally pay validators to seed so you have backup and sync when the desktop is off.
3. **Artists pay for catalog replication** — catalog content is where validators get paid to replicate and distribute; personal library is your devices first, validators optional.
4. **Scalability** — deduplicate when FLAC CID matches (reference catalog or one blob, many owners); small replication for library; optional convergent encryption.
5. **Physical → digital** — redeem codes for real “own CD/vinyl → get digital”; same FLAC as catalog (CD) as honor-system entitlement.
6. **Future-proof** — transfer of consumer entitlement (trades, secondhand sales), optional acoustic match (local AI) for association/dedup.

The protocol stays one stack (consensus, storage, encryption, policy); we add **content source** (catalog vs personal library), **device sync** (your devices first, validators when you pay), and reserve **transfer of consumer rights** and optional **acoustic match** for later. Docs and UX shift to “library first, distribution is one way to fill it.”
