# Economics

Mojave has **no native token**. Everyone is identified by an **Ed25519 public key**. Payments use **USDC attestations** for (1) user subscriptions (library size) and (2) content purchases. Validators take a cut of both; artist sale volume and user subscription fees pay for hosting. Users and good samaritans seed content via BitTorrent; validators are not the only source of bytes.

## Identity

| Concept | Implementation |
|--------|-----------------|
| Identity | Ed25519 public key. No on-chain "account balance" in a native token. |
| Users | Have a pubkey. Pay for subscription (library size) via USDC attestation. |
| Artists / distributors | Sign up as a distributor. Grant access to content when they accept a USDC purchase attestation. |
| Validators | Elected (see [governance.md](governance.md)). Earn from cut of USDC flows (subscriptions + sales). |

## User subscriptions (library size)

Users who want validators to replicate their **personal library** (backup, sync when desktop is off) pay a **subscription** based on library size. Payment is via **USDC attestation** (e.g. user paid $X for tier Y or per GB). Validators (or an attestation verifier) accept the attestation; the user's library is then eligible for replication up to the subscribed tier. Subscription revenue flows to validators and pays for that hosting.

- **Attestation:** Off-chain or cross-chain proof that "pubkey X paid $Y for subscription Z (e.g. 50 GB)." Verifier checks it and allows the user to use validator-backed replication for their library.
- **Users are also seeders:** Beyond subscription, users seed their own content (e.g. Tauri desktop seeds; phone leeches from desktop). Good samaritans can seed too. So the system is like BitTorrent: validators are one source of bytes, not the only one; subscription pays for the validator-hosted slice.

## Content purchases (artist sales)

Artists **sign up as distributors**. They publish content and set prices (in USD/USDC). When a user wants to buy access:

1. User pays in USDC (via a payment rail that produces an attestation).
2. **Attestation** proves "pubkey X paid $Y to distributor Z for content C" (or for a bundle).
3. Artist (distributor) or a validator acting on their behalf **grants access** to the user: issues a wrapped DEK (or records an entitlement) so the user can decrypt. Grant is gated on a valid purchase attestation.
4. **Validators take a cut** of the sale (e.g. a percentage). The rest goes to the artist. Artist sale volume thus pays for artist hosting — validators earn from the cut and use that to cover the cost of storing and serving that artist's content.

So: **artist sale volume + user subscription fees** fund validator hosting. No MOJ; all payments in USDC with attestations proving payment.

## Validator revenue

Validators earn from:

1. **Cut of content purchases (USDC).** When a user buys access from an artist, validators take a defined cut; the artist receives the remainder. This pays for hosting artist content.
2. **User subscription fees (USDC).** Users pay subscriptions (by library size) via attestation; that revenue goes to validators and pays for user library backup/replication.

No block rewards, no gas fees in a native token. Validator economics are **USDC in, hosting out**. Slashing / removal for misbehavior can still exist (e.g. stake in a separate asset, or reputation/recall only); see [governance.md](governance.md).

## Replication and seeding

- **Validators** replicate content they are paid to host (artist catalog + subscribed user libraries). They seed via BitTorrent like everyone else.
- **Users** seed their own library from their desktop (and optionally other devices). They are their own seeders.
- **Good samaritans** can seed any content they have (ciphertext only); no payment. Same as BitTorrent: voluntary seeders improve availability.

So validators are not the only seeders; they are the **paid** seeders for catalog and subscribed user content. The rest is P2P.

## Attestations (summary)

| Flow | Attestation | Who verifies | Outcome |
|------|-------------|--------------|---------|
| User subscription | "Pubkey X paid $Y for subscription tier Z (library size)" | Validator or attestation service | User gets validator-backed replication for their library up to tier. |
| Content purchase | "Pubkey X paid $Y to distributor Z for content C" | Artist (distributor) or validator | User gets access (entitlement + wrapped DEK). Validator takes cut; artist gets rest. |

Attestation format and verification (on-chain vs off-chain, which chain or payment rail) are TBD. The design assumption is that attestations are verifiable and binding so that (1) subscriptions can be metered and (2) artists can grant access on proof of payment.

## What is not in scope

- **No MOJ, no grains, no native token.** No gas fees in a protocol token. No staking in a protocol token (staking for validator admission, if any, is separate — see governance).
- **No on-chain USDC transfer.** USDC moves on its own rail (e.g. Ethereum, Solana, CEX). The protocol only consumes **attestations** that a payment occurred. Settlement is off-chain or on the chain where USDC lives.
- **No block rewards / inflation.** Validators are compensated from USDC flows only.

## Chain parameters (governance-adjustable)

Parameters that may still apply (spam prevention, policy execution):

| Parameter | Purpose |
|-----------|---------|
| `max_script_gas` | Gas budget for Goja script execution (policy / proofs). Can be a simple cap; no token fee. |
| Rate limits | Transaction submission rate limits per pubkey to prevent spam (no fee, just throttle). |

Validator election parameters (stake threshold if any, voting rules) are in [governance.md](governance.md). Revenue share (validator cut of sales, subscription split) is a governance or contractual matter.
