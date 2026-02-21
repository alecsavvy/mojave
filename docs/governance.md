# Governance

> **[Governance overview diagram](diagrams/governance.mermaid)** · **[Takedown flow diagram](diagrams/takedown.mermaid)**

This document covers validator elections, oracle elections, copyright takedowns, jurisdictional compliance, and the on-chain governance process that ties them together.

## Why governance matters here

Most blockchains treat validator selection as a purely economic question — stake enough tokens, you're in. That works for general-purpose chains where validators are fungible compute providers. It doesn't work for a music distribution network.

Mojave validators handle unencrypted audio during processing, hold DEKs that grant access to protected content, and serve as the trust layer between rights holders and consumers. The people operating these nodes need to be accountable — not just economically, but socially. A rights holder shipping their catalog to a network of anonymous stakers is a non-starter for the music industry.

## Validator elections

Validator admission is a social election, not just a staking threshold. Staking is necessary but not sufficient. The network's users — artists, labels, consumers, existing validators — vote on who gets to operate a node.

### Election process

1. **Candidacy.** An entity announces candidacy on-chain via a `SubmitCandidacy` transaction. The candidacy includes:
   - Stake deposit (minimum threshold, governance-set).
   - Identity disclosure — who they are, where they operate, their jurisdiction(s).
   - Infrastructure commitment — hardware specs, uptime SLA, geographic location.
   - Motivation statement — why they want to operate a node (this is social, not technical).

2. **Voting period.** A governance-set window (e.g. 14 days) during which token holders vote. Voting power is proportional to staked MOJ (including delegated stake). Votes are `yes`, `no`, or `abstain`.

3. **Threshold.** Candidacy passes if it exceeds a quorum (e.g. 33% of staked tokens participate) and a supermajority (e.g. 67% of votes are `yes`). These parameters are governance-adjustable.

4. **Admission.** On passing, the candidate's validator is added to the active set at the next epoch boundary. They begin participating in consensus, receiving block rewards, and accepting replication/DEK holder assignments.

### Removal

Validators can be removed by the same election mechanism in reverse:

1. **Recall proposal.** Any token holder can submit a `RecallValidator` proposal, citing reasons (misconduct, downtime, trust violation, etc.).
2. **Voting period.** Same window and thresholds as admission.
3. **Removal.** On passing, the validator enters the standard unbonding period. Their stake is locked during unbonding and subject to slashing if misconduct is proven during that window. They stop participating in consensus immediately.

Automatic removal (slashing for double-signing, extended downtime) still applies as a backstop — elections handle the social layer, slashing handles the protocol layer.

### Why this works for music

The election process means the network can vote in entities the industry actually trusts — a major label's infrastructure arm, a well-known indie distributor, a music tech company with a track record. It can also block entities the community doesn't trust, regardless of how much they're willing to stake.

This is different from anonymous DeFi validation. The music industry needs to know who is handling their content. Elections create that accountability without requiring a centralized authority to grant permission.

| Validator type | Example | Why the network might vote them in |
|---------------|---------|-----------------------------------|
| Major label infrastructure | Sony Music's cloud ops team | Industry trust, existing relationships, legal accountability |
| Indie distributor | DistroKid, CD Baby | Already trusted by millions of independent artists |
| Music tech company | A well-known music API provider | Technical credibility, existing infrastructure |
| Community operator | A respected community member with proven ops track record | Grassroots trust, geographic diversity |
| Academic / nonprofit | A university research lab, Internet Archive | Neutrality, long-term commitment, no commercial conflict |

## Oracle elections

Oracles are elected entities with a specific authority: they can submit copyright takedown requests on-chain. Oracles are the bridge between the legal system and the protocol. They don't have validator powers — they can't participate in consensus, produce blocks, or hold DEKs. Their only capability is submitting `TakedownRequest` transactions.

### Why oracles, not validators

Validators are infrastructure operators. Copyright enforcement is a legal and social function — it requires understanding of copyright law, DMCA processes, rights databases, and dispute resolution. Conflating the two roles creates perverse incentives: a validator that's also a copyright enforcer might selectively take down competitors' content, or be pressured by a rights holder to favor their takedowns.

Separating the roles means:
- Validators focus on infrastructure — seeding, DEK management, consensus.
- Oracles focus on copyright — verifying claims, submitting takedowns, handling disputes.
- Neither can interfere with the other's domain.

### Election process

Oracles go through the same election process as validators — candidacy, voting, threshold, admission. The candidacy includes additional requirements:

- Demonstrated expertise in copyright law or rights management.
- Jurisdiction(s) they cover.
- Dispute resolution process they commit to following.
- Contact information for counter-notices.

Oracle elections are **per-network** — different Mojave deployments (e.g. a US-focused network vs. a global network vs. a region-specific network) can elect different oracle sets appropriate to their jurisdiction and community.

### Oracle authority

An elected oracle can submit:

| Transaction | Effect |
|-------------|--------|
| `TakedownRequest` | Initiates a takedown process for a specific CID |
| `TakedownResolve` | Resolves a takedown — either confirms removal or dismisses the claim |

Oracles cannot directly delete content or revoke DEKs. They submit requests that the protocol processes through a defined workflow (see Takedowns below).

## Takedowns

When content needs to be removed — copyright infringement, illegal material, court order — the takedown mechanism is DEK removal. The encrypted `.flac.tdf` blobs become permanently inaccessible without destroying the audit trail.

### Why DEK removal, not file deletion

Deleting files from a BitTorrent swarm is practically impossible — anyone who already has a copy can re-seed it. But without the DEK, the file is useless ciphertext. DEK removal is the effective kill switch:

1. Wrapped DEKs are removed from the chain store (uploader's on-chain wrapped DEK) and from all validators' local stores.
2. The access policy for the CID is set to `TAKEN_DOWN` — a terminal state.
3. Validators stop issuing `GrantAccess` for this CID.
4. Existing wrapped DEKs on consumer devices expire normally and cannot be refreshed.
5. The `.flac.tdf` files may persist in the BitTorrent swarm, but they're undecryptable.

The on-chain records remain — the upload, the release, the access grants, the takedown itself. This is the global paper trail. Every action is signed and timestamped. Who uploaded it, who claimed it, who took it down, why, when. Immutable audit history.

### Takedown flow

```
Claimant                Oracle              Chain              Validators
  |                       |                   |                     |
  |-- copyright claim --->|                   |                     |
  |   (off-chain:         |                   |                     |
  |    evidence, CID,     |                   |                     |
  |    legal basis)        |                   |                     |
  |                       |                   |                     |
  |             oracle reviews claim           |                     |
  |             verifies evidence              |                     |
  |                       |                   |                     |
  |                       |-- TakedownRequest -->|                     |
  |                       |   (CID, oracle sig, |                     |
  |                       |    reason, evidence |                     |
  |                       |    hash)            |                     |
  |                       |                   |                     |
  |                       |                   |-- takedown.pending -->|
  |                       |                   |   event               |
  |                       |                   |                     |
  |              counter-notice window         |                     |
  |              (governance-set, e.g. 14 days)|                     |
  |                       |                   |                     |
  |  (if no counter-notice or counter-notice rejected)               |
  |                       |                   |                     |
  |                       |-- TakedownResolve -->|                     |
  |                       |   (confirmed)      |                     |
  |                       |                   |                     |
  |                       |                   |-- takedown.confirmed ->|
  |                       |                   |                     |
  |                       |                   |  delete uploader's    |
  |                       |                   |  on-chain wrapped DEK |
  |                       |                   |                     |
  |                       |                   |                     |-- delete local
  |                       |                   |                     |   wrapped DEKs
  |                       |                   |                     |-- stop seeding
  |                       |                   |                     |   (optional)
  |                       |                   |                     |
  |                       |                   |  access policy →     |
  |                       |                   |  TAKEN_DOWN          |
  |                       |                   |                     |
```

### Counter-notices

The uploader (or content owner) can submit a counter-notice during the window, claiming the takedown is invalid. The counter-notice is on-chain — signed, timestamped, part of the audit trail.

If a counter-notice is submitted:
1. The oracle reviews the counter-notice and any supporting evidence.
2. The oracle submits `TakedownResolve` with a decision — confirmed (takedown stands) or dismissed (content is restored).
3. If dismissed: the `TAKEDOWN_PENDING` status is removed, the content remains accessible, and the takedown request stays on-chain as a record.

If the oracle's decision is disputed, the community can recall the oracle via the election mechanism. The protocol doesn't adjudicate copyright disputes — it provides the process and the paper trail. Final legal disputes happen off-chain in courts.

### Takedown state

A new access policy state is added:

| Status | Meaning |
|--------|---------|
| `TAKEDOWN_PENDING` | A takedown has been requested. Content is still accessible during the counter-notice window. |
| `TAKEN_DOWN` | Takedown confirmed. DEKs removed. Content is cryptographically inaccessible. |

Both states are on-chain and queryable. A UI built on the API can show takedown status to users.

### What takedowns don't solve

This system doesn't solve DRM or piracy any more than Spotify, Apple Music, or any other platform can. Once audio is decrypted and playing through speakers, it can be recorded. Once a consumer has a valid wrapped DEK, they can decrypt and redistribute the FLAC. This is true of every digital distribution system.

What the on-chain audit trail does provide:

- **Deterrence.** Every access grant is signed and timestamped. If leaked content is traced back to a specific access grant, the consumer who leaked it is identifiable by their Ed25519 key. This doesn't prevent leaking, but it creates accountability.
- **Proof of infringement.** The takedown itself — who filed it, when, what evidence, what CID — is permanently on-chain. Useful for legal proceedings.
- **Global consistency.** A takedown on one validator is a takedown on all validators. No fragmented enforcement across jurisdictions (at the protocol level — jurisdictional content filtering is handled separately, see below).

## Jurisdictional compliance

Validators operate in specific legal jurisdictions. Content that's legal in one jurisdiction may be illegal in another. The protocol needs a mechanism for validators to comply with local law without requiring global consensus on what's legal where.

### Jurisdiction declarations

Each validator declares its operating jurisdiction(s) in its candidacy and on-chain profile. This is a self-reported list of ISO 3166-1 country codes. The declaration is public — anyone can see where a validator claims to operate.

### Jurisdictional content filtering

Validators can locally refuse to store or serve content that violates their jurisdiction's laws. This is a **local policy decision**, not a consensus action:

1. A validator receives an `upload.complete` event for content it's in the replication set for.
2. Before fetching, it checks the content's metadata (DDEX territory restrictions, takedown status, flagged content lists) against its jurisdictional policy.
3. If the content is prohibited in its jurisdiction: the validator does not fetch or seed it. It remains in the replication set on-chain (other validators in legal jurisdictions serve it), but this validator opts out locally.

This is functionally equivalent to how CDNs handle jurisdictional content — a CDN edge node in Germany might not serve content that's legal in the US. The content still exists on the network; it's just not served from that location.

### Jurisdictional oracle sets

Because oracle elections are per-network, different deployments can have oracle sets appropriate to their jurisdiction:

- A US-focused deployment elects oracles familiar with DMCA.
- An EU-focused deployment elects oracles familiar with the DSA and local copyright directives.
- A global deployment might elect oracles from multiple jurisdictions with defined handoff procedures.

The protocol doesn't encode any specific copyright law. It provides the mechanism (elected oracles, takedown workflow, counter-notices, audit trail) and lets each deployment configure the specifics.

### Content flagging

In addition to full takedowns, oracles can flag content with jurisdictional restrictions:

| Transaction | Effect |
|-------------|--------|
| `FlagContent` | Adds a jurisdictional flag to a CID — e.g. "restricted in DE, FR" |
| `UnflagContent` | Removes a jurisdictional flag |

Flags don't remove DEKs or block access globally. They're advisory metadata that validators and clients can use to comply with local law. A validator in Germany that sees a `restricted in DE` flag can stop serving the content locally. A validator in the US ignores the flag.

Flags are on-chain — transparent, auditable, queryable. A UI can filter content based on the user's locale and active flags.

## Governance proposals

Beyond elections, the network uses on-chain governance proposals for protocol-level decisions. These follow standard Cosmos SDK governance patterns:

### Proposal types

| Proposal | Purpose | Threshold |
|----------|---------|-----------|
| `ValidatorCandidacy` | Elect a new validator | Supermajority of staked votes |
| `RecallValidator` | Remove a validator | Supermajority of staked votes |
| `OracleCandidacy` | Elect a new oracle | Supermajority of staked votes |
| `RecallOracle` | Remove an oracle | Supermajority of staked votes |
| `ParameterChange` | Modify chain parameters (gas prices, inflation, etc.) | Supermajority of staked votes |
| `SoftwareUpgrade` | Coordinate a chain upgrade | Supermajority of staked votes |
| `CommunitySpend` | Allocate funds from the community/treasury pool | Supermajority of staked votes |

### Voting

- Voting power is proportional to staked MOJ.
- Delegators inherit their validator's vote by default but can override.
- Voting period is governance-set (default 14 days).
- Quorum: 33% of staked tokens must vote.
- Supermajority: 67% of votes must be `yes`.
- Veto threshold: if > 33% of votes are `no_with_veto`, the proposal fails regardless of yes votes and the proposer's deposit is burned.

### Governance parameters

| Parameter | Default | Controls |
|-----------|---------|----------|
| `voting_period` | 14 days | Duration of the voting window |
| `quorum` | 33% | Minimum participation for a valid vote |
| `threshold` | 67% | Minimum yes votes to pass |
| `veto_threshold` | 33% | Minimum veto votes to kill a proposal |
| `min_deposit` | 1,000 MOJ | Minimum deposit to submit a proposal |
| `counter_notice_window` | 14 days | Time for content owners to respond to a takedown |
| `max_oracles` | 21 | Maximum number of active oracles |
| `max_validators` | 100 | Maximum number of active validators |
