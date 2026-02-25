# Governance

> **[Governance overview diagram](diagrams/governance.mermaid)** · **[Takedown flow diagram](diagrams/takedown.mermaid)**

This document covers validator elections, publisher groups (“states”), publisher admission, copyright takedowns (by each group’s designated takedown authority), jurisdictional compliance, federal governance with group representation, and how clients and indexers work across multiple chains.

The design is inspired by a **federal/state** split: the chain is the federal layer (validators run neutral infrastructure and serve all groups equally); **publisher groups** are like states, each with their own upload policy and **takedown authority** — the group’s development or compliance arm designates who can submit takedowns for that group’s content. There are no network-wide oracles; compliance and takedowns are entirely in the hands of each publisher group. Validators have the most scrutiny on admitting new groups. Federal governance gives groups representation based on activity (and optionally a bond or stake in a non-protocol asset; the protocol has no native token — see [economics.md](economics.md)).

## Why governance matters here

Most blockchains treat validator selection as a purely economic question — meet an economic threshold, you're in. That works for general-purpose chains where validators are fungible compute providers. It doesn't work for a music distribution network.

Mojave validators handle unencrypted audio during processing, hold DEKs that grant access to protected content, and serve as the trust layer between rights holders and consumers. The people operating these nodes need to be accountable — not just economically, but socially. A rights holder shipping their catalog to a network of anonymous operators is a non-starter for the music industry.

## Validator elections

Validator admission is a **social election**. The network's users — artists, labels, consumers, existing validators — vote on who gets to operate a node. Validators earn from **USDC flows** (cut of content sales + user subscription fees), not from a native token. Stake or other economic criteria for candidacy/voting are TBD (see [economics.md](economics.md)).

### Election process

1. **Candidacy.** An entity announces candidacy on-chain via a `SubmitCandidacy` transaction. The candidacy includes:
   - Identity disclosure — who they are, where they operate, their jurisdiction(s).
   - Infrastructure commitment — hardware specs, uptime SLA, geographic location.
   - Motivation statement — why they want to operate a node (this is social, not technical).
   - (Optional) Stake or bond — governance may require a deposit or bond; TBD.

2. **Voting period.** A governance-set window (e.g. 14 days) during which eligible participants vote. Voting power and eligibility are governance-set (e.g. one vote per pubkey, or weighted by activity; no MOJ). Votes are `yes`, `no`, or `abstain`.

3. **Threshold.** Candidacy passes if it exceeds a quorum and a supermajority. These parameters are governance-adjustable.

4. **Admission.** On passing, the candidate's validator is added to the active set at the next epoch boundary. They begin participating in consensus and accepting replication/DEK holder assignments. They earn from USDC (cut of sales + subscriptions); no block rewards in a native token.

### Removal

Validators can be removed by the same election mechanism in reverse:

1. **Recall proposal.** Any eligible participant can submit a `RecallValidator` proposal, citing reasons (misconduct, downtime, trust violation, etc.).
2. **Voting period.** Same window and thresholds as admission.
3. **Removal.** On passing, the validator is removed from the active set. Unbonding period and slashing (if any) are governance-set; no native token required.

Automatic removal (slashing for double-signing, extended downtime) still applies as a backstop — elections handle the social layer, slashing handles the protocol layer.

### Why this works for music

The election process means the network can vote in entities the industry actually trusts — a major label's infrastructure arm, a well-known indie distributor, a music tech company with a track record. It can also block entities the community doesn't trust, regardless of economic weight.

This is different from anonymous DeFi validation. The music industry needs to know who is handling their content. Elections create that accountability without requiring a centralized authority to grant permission.

| Validator type | Example | Why the network might vote them in |
|---------------|---------|-----------------------------------|
| Major label infrastructure | Sony Music's cloud ops team | Industry trust, existing relationships, legal accountability |
| Indie distributor | DistroKid, CD Baby | Already trusted by millions of independent artists |
| Music tech company | A well-known music API provider | Technical credibility, existing infrastructure |
| Community operator | A respected community member with proven ops track record | Grassroots trust, geographic diversity |
| Academic / nonprofit | A university research lab, Internet Archive | Neutrality, long-term commitment, no commercial conflict |

### Validators as federal infrastructure

Validators are the **federal** layer. They run neutral infrastructure and must serve **all** admitted publisher groups equally. A validator cannot refuse to serve one group (e.g. Audius) while serving another (e.g. Sony). Replication and DEK holder assignments apply to content from every group; validators do not get to opt out by group. Jurisdictional content filtering (see below) remains the only exception — a validator may locally refuse content that violates its declared legal jurisdiction (e.g. not serve in Germany content that is illegal there). That is a legal compliance carve-out, not a choice to discriminate between groups.

Validators also have **the most scrutiny on allowing new states**. Admitting a new publisher group (a new “state” on the chain) requires a governance process in which the validator set has the primary or heaviest weight. So creating a new group is not permissionless; it is validator-scrutinized. See Publisher groups and Governance proposals below.

## Publisher groups (states)

Publisher groups are the **state** layer. Each group (e.g. a label, Audius, Sony, an indie collective) has its own “laws”: who can publish under it and who can initiate takedowns for its content. Compliance and takedowns are **entirely in the hands of the publisher group’s development or compliance arm** — there are no network-wide oracles. Content is always associated with a publisher group (or marked independent). Only that group’s designated **takedown authority** can submit a takedown for that group’s content; Sony cannot file a takedown for content under Audius’ group. This prevents one group from bullying another’s content.

### Group identity and on-chain state

- Each **group** has an on-chain identity: `group_id` (or group account/pubkey). The chain maintains a group registry: which groups exist, their bond or stake (if any; TBD, no protocol token), and the group’s **takedown authority** — the pubkey(s) that can submit takedowns (e.g. DDEX purge/takedown release messages) and resolve them for that group’s content. The group (e.g. its development arm, legal, or compliance team) controls this key; no separate oracle election is needed.
- **Content** (each upload/release) is tagged with a **publisher_group_id**. So the chain can answer “which group backs this CID?”
- **Who can publish under a group** is determined by that group: an allow-list of pubkeys, or a role/capability the group key signs. Groups manage this off-chain or via on-chain rules they control; the chain only checks at upload time that the uploader is allowed for the claimed group (or meets chain-wide rules for “independent” uploads).

### Creating a new group (admitting a new state)

Creating a new publisher group is **validator-scrutinized**. It is not permissionless. A proposal to admit a new group (e.g. `AdmitGroup`) is decided by governance in which **validators have the most say** — e.g. validator-weighted vote, or a dual requirement (governance-set majority and validator supermajority). This mirrors the high bar for “adding a new state” in a federal system. The proposal typically includes:

- Group identity and **designated takedown authority** (the pubkey(s) that will submit takedowns for this group’s content — e.g. the group’s compliance or development arm).
- Optional: minimum bond or stake the group will lock (governance can require a bond and slashing on takedown; no protocol token — see economics.md).
- Contact and dispute-resolution commitment.

Once admitted, the group appears in the group registry and can receive content tagged with its `group_id`. Only that group’s takedown authority can submit a takedown for that group’s content — using **DDEX’s takedown/purge release message** (or an on-chain transaction that carries it), so groups use the same industry-standard signal for “remove this release.” The group can update its takedown authority via group governance or a designated update path (e.g. group key signs a `SetGroupTakedownAuthority` transaction).

### Group bond and slashing (optional)

A chain may require groups to post a **bond** (or allow optional bonding); there is no protocol native token. If content under a group receives a **confirmed takedown**, that group’s bond can be **slashed** (fully or partially). That ties abuse (copyright infringement, illegal content) to economic cost for the group and discourages groups from turning a blind eye. Parameters (slash amount, unbonding) are governance-set.

### Relation to publisher admission

Publisher admission (who can publish on the chain) is now **group-aware**. At upload time the chain checks: (1) the claimed publisher group exists and (if required) has sufficient bond or meets criteria; (2) the uploader is allowed to publish under that group (per the group’s allow-list or rules); and (3) any chain-wide floor (e.g. “all publishers must be in a group or meet independent criteria”) is satisfied. So the chain sets the floor; groups are the primary unit of “who can publish,” and each group has its own culture (e.g. Audius more permissive, Sony more bureaucratic).

## Publisher admission

Who can publish content on a chain is determined by **chain-level policy** and **publisher groups** (see above). You publish **under a group** (or as independent if the chain allows it). The chain maintains a floor (e.g. “must be in an admitted group” or “must meet chain-wide Casbin rules”); each group defines who can publish under it (allow-list, roles, or off-chain process). The same on-chain policy machinery used for access control (Casbin models and rules, optionally Goja scripts) can be used for chain-wide publisher admission and for group-level rules — evaluated at upload time.

### How it works

- **Group-aware upload.** Every upload/release specifies a **publisher_group_id** (or “independent”). The chain checks: (1) the group exists (and meets required criteria if applicable); (2) the uploader is allowed to publish under that group; (3) any chain-wide publisher policy is satisfied. If any check fails, the validator rejects the upload.
- **On-chain policy.** The chain may maintain a **chain-wide publisher policy** (Casbin model + rules, or Goja script) as a floor — e.g. “independent publishers must meet these rules” or “all publishers must belong to a group.” Group membership and per-group allow-lists are stored on-chain or proven at upload time. The model ID for publisher admission is a governance parameter.
- **Governance.** Governance proposals (e.g. `PublisherPolicyChange`, `AdmitGroup`) update chain-wide policy and admit new groups. Validators enforce whatever is on-chain and do not unilaterally decide who can publish.

### Relation to takedowns

- **Publisher admission** is **proactive**: it decides who can create new content and under which group.
- **Takedowns** are **reactive** and **group-scoped**: only the content’s publisher group’s **takedown authority** (the pubkey(s) designated by that group) can submit a takedown for that content. Content that passed publisher admission can still be taken down by that group’s takedown authority if the group decides a valid claim has been made.

## Takedowns

When content needs to be removed — copyright infringement, illegal material, court order — the takedown mechanism is DEK removal. **Publisher groups initiate takedowns** by sending the same kind of signal the industry already uses: **DDEX’s takedown/purge release message**. Content that slipped under a group’s radar (e.g. infringing uploads, court-ordered removal) is taken down when that group’s takedown authority submits a purge/takedown release message (or an on-chain transaction that carries it). The chain then runs the takedown workflow (counter-notice window, DEK removal). **Before any workflow runs**, the chain checks that the signer is the **takedown authority** for the CID’s publisher group (i.e. one of the pubkeys registered for that group in the group registry). If not, the request is rejected. This applies to any content that has been published on the chain. The encrypted `.flac.tdf` blobs become permanently inaccessible without destroying the audit trail. Compliance and the decision to take down content are entirely in the hands of each publisher group’s designated authority (e.g. their development or legal arm).

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
Claimant         Group's takedown authority   Chain              Validators
  |                       |                   |                     |
  |-- copyright claim --->|                   |                     |
  |   (off-chain:         |                   |                     |
  |    evidence, CID,     |                   |                     |
  |    legal basis)       |                   |                     |
  |                       |                   |                     |
  |             group's authority reviews      |                     |
  |             claim, verifies evidence       |                     |
  |                       |                   |                     |
  |                       |-- TakedownRequest -->|                     |
  |                       |   (CID, sig,       |                     |
  |                       |    reason, evidence|                     |
  |                       |    hash)           |                     |
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
1. The group’s takedown authority reviews the counter-notice and any supporting evidence.
2. The takedown authority submits `TakedownResolve` with a decision — confirmed (takedown stands) or dismissed (content is restored).
3. If dismissed: the `TAKEDOWN_PENDING` status is removed, the content remains accessible, and the takedown request stays on-chain as a record.

If the group’s decision is disputed, the uploader or community can pursue the group (e.g. via group governance, or off-chain). The protocol doesn’t adjudicate copyright disputes — it provides the process and the paper trail. Final legal disputes happen off-chain in courts.

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

The protocol doesn't encode any specific copyright law. It provides the mechanism (per-group takedown authority, takedown workflow, counter-notices, audit trail); each group runs its own compliance.

### Content flagging

In addition to full takedowns, a group’s **takedown authority** can flag that group’s content with jurisdictional restrictions:

| Transaction | Effect |
|-------------|--------|
| `FlagContent` | Adds a jurisdictional flag to a CID. **Accepted only if** the signer is the takedown authority for that CID’s publisher group. E.g. "restricted in DE, FR". |
| `UnflagContent` | Removes a jurisdictional flag (same authority check). |

Flags don't remove DEKs or block access globally. They're advisory metadata that validators and clients can use to comply with local law. A validator in Germany that sees a `restricted in DE` flag can stop serving the content locally. A validator in the US ignores the flag.

Flags are on-chain — transparent, auditable, queryable. A UI can filter content based on the user's locale and active flags.

## Clients and indexers

Chains are independent. There is no protocol-level requirement for chains to talk to each other (no IBC). **Identity is portable**: the same Ed25519 keypair (wallet) works on every chain. What changes is which chain or chains the client talks to.

- **Clients** (music players, storefronts, upload tools) are configured with the chain or chains relevant to them — e.g. a single indie-chain API, or a list of chain endpoints (indie, major-label, permissionless). The user’s address is the same on all of them; the client switches context the way a wallet switches from Ethereum mainnet to Sepolia. The client discovers content (which may live on different chains), then uses the appropriate chain’s API for upload, access, or purchase.
- **Indexers** (search, catalog, discovery) pull from the network(s) they care about. They may index one chain, or many, and expose a unified or filtered view. Discovery (“Green Day is on chain B”) can be off-chain (registry, config) or derived from indexed chain data. No cross-chain proofs are required — just pointing at the right chain’s state and API.

So “interconnect” is client- and indexer-side: they pull from the relevant networks. Each chain’s state (including publisher policy, entitlements, and takedowns) is self-contained; clients and indexers aggregate as needed.

## Federal governance and group representation

Federal (chain-wide) governance decides parameter changes, new group admission, validator elections, and upgrades. To avoid pure plutocracy and to give **publisher groups** a voice, the chain can use **dual representation**: (1) **stake-weighted** voting (existing: token holders vote with staked MOJ), and (2) **group representation** where each admitted group gets a vote weight based on **stake + activity** for that group.

### Group vote weight

Each group’s voting power in federal governance is a function of:

- **Stake** — MOJ staked or locked by/for that group (or attributed to the group in the registry). Reflects skin in the game.
- **Activity** — an on-chain measure of usage over a recent window, e.g. number of CIDs published under that group, access grants for that group’s content, or similar. Prevents inactive capital from dominating and rewards real use.

The exact formula (e.g. weighted sum or product of stake and activity, with caps) is governance-set. Newly admitted groups may have no activity yet; they can be given a default weight or required to hold minimum stake to participate. Activity metrics should be costly to game (e.g. tied to upload/storage fees) and computed over a rolling window.

### How group and stake votes combine

Options (governance chooses):

- **Dual majority:** A proposal passes only if it gets a supermajority of **stake-weighted** votes **and** a supermajority of **group-representation** votes. So both token holders and groups must agree.
- **Blended weight:** Total voting power = α × (stake-weighted total) + β × (group-representation total), with α, β governance parameters. A single roll call with combined weight.
- **Proposal-type specific:** e.g. new group admission (`AdmitGroup`) requires validator supermajority or group+stake dual majority; parameter changes might be stake-only.

Validators have **the most scrutiny on admitting new groups**: e.g. `AdmitGroup` proposals require validator-weighted supermajority (or a process where validators’ vote is the primary gate). Group representation then gives admitted groups a say in subsequent federal decisions (parameters, validator elections, etc.) based on their stake and activity.

## Governance proposals

Beyond elections, the network uses on-chain governance proposals for protocol-level decisions. Federal (chain-wide) proposals can use both **stake-weighted** voting and **group representation** (see above). Admitting a new publisher group is validator-scrutinized.

### Proposal types

| Proposal | Purpose | Threshold |
|----------|---------|-----------|
| `AdmitGroup` | Admit a new publisher group (“new state”). Requires **validator-scrutinized** approval (e.g. validator supermajority or dual stake + group majority). | Validator supermajority or governance-set dual majority |
| `ValidatorCandidacy` | Elect a new validator | Supermajority of staked votes (and optionally group representation) |
| `RecallValidator` | Remove a validator | Supermajority of staked votes |
| `PublisherPolicyChange` | Update chain-wide publisher admission (model, rules, or script reference) | Supermajority of staked votes |
| `ParameterChange` | Modify chain parameters (gas prices, inflation, publisher policy model ID, group vote formula, etc.) | Supermajority of staked votes |
| `SoftwareUpgrade` | Coordinate a chain upgrade | Supermajority of staked votes |
| `CommunitySpend` | Allocate funds from the community/treasury pool | Supermajority of staked votes |

### Voting

- **Stake-weighted voting:** Voting power is proportional to staked MOJ. Delegators inherit their validator's vote by default but can override.
- **Group representation (optional per proposal type):** Each admitted group has a vote weight = f(group stake, group activity). The group key (or designated signer) casts the group’s vote. Proposals can require both a stake majority and a group-representation majority, or use a blended weight; see Federal governance and group representation.
- Voting period is governance-set (default 14 days).
- Quorum: 33% of participating stake (and, if applicable, quorum of group representation) must vote.
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
| `publisher_policy_model_id` | (governance-set) | Casbin model ID for chain-wide publisher admission; rules are updated via `PublisherPolicyChange` |
| `group_vote_activity_window` | 180 days | Rolling window (e.g. blocks or time) for computing group activity in federal vote weight |
| `group_vote_stake_weight` | (governance-set) | Weight α for stake in group vote weight formula (e.g. weight = α×sqrt(stake) + β×sqrt(activity)) |
| `group_vote_activity_weight` | (governance-set) | Weight β for activity in group vote weight formula |
| `max_validators` | 100 | Maximum number of active validators |
| `min_group_stake` | (governance-set) | Minimum stake (if any) for a group to be admitted or to participate in group-representation voting |
