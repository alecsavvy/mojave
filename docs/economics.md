# Economics

Mojave has a single native token used for gas fees, storage fees, content purchases, and validator rewards. The token follows a dual-denomination model inspired by Solana's SOL/lamports split.

## Token

| | |
|---|---|
| **Token** | MOJ (human-readable unit) |
| **Base unit** | grain (smallest indivisible unit) |
| **Conversion** | 1 MOJ = 1,000,000,000 grains (10^9) |

All on-chain values are denominated in grains. MOJ is a display-layer convenience — the chain never sees it. This is the same relationship as SOL/lamports or ETH/wei.

Grains exist so gas fees can be extremely small numbers without requiring fractional tokens. A transaction that costs 5,000 grains is 0.000005 MOJ — effectively free from the user's perspective, but nonzero enough to prevent spam.

## Fee model

Every transaction has a fee. Fees serve two purposes: spam prevention and validator compensation.

### Gas fees (all transactions)

Every transaction pays a base gas fee proportional to its computational cost. Gas is priced in grains per unit, with a minimum floor set by chain governance.

| Transaction | Relative cost | Why |
|-------------|--------------|-----|
| `PublishRelease` | Low | Small state write — CID + metadata reference |
| `GrantAccess` | Low | Small state write — audit log entry |
| `AddPolicy` / `RemovePolicy` | Low | Small state write — Casbin rule |
| `SetAccessPolicy` | Low | Small state write — policy reference |
| `DelegateRole` / `RevokeEntitlement` | Low | Small state write — entitlement |
| `DeployScript` | Medium | Stores script source on-chain — sized proportionally |
| `SubmitAttestation` | Medium | Stores attestation result on-chain |
| `UploadComplete` | High | Triggers replication, DEK distribution, event emission |
| `RotateDEK` | High | Re-encryption, re-distribution, re-seeding |
| `RecoverDEKHolders` | Medium | DEK re-wrapping + distribution |

Gas prices float within a governance-set range. Validators include transactions ordered by fee (standard mempool priority). During congestion, fees rise naturally. During idle periods, fees sit at the floor.

### Storage fees

Uploading content imposes an ongoing cost on the network — validators store and seed files indefinitely. Storage fees are a one-time payment at upload, scaled by file size, that compensates the replication set for their commitment.

| Component | Fee basis |
|-----------|----------|
| Audio (`.flac.tdf`) | Per-byte of the encrypted blob |
| Images (`.png`) | Per-byte, per image |
| Replication factor | Multiplied by the number of validators in the replication set |

```
storage_fee = (audio_bytes + sum(image_bytes)) × rate_per_byte × replication_set_size
```

The `rate_per_byte` is a chain parameter, adjustable by governance. A larger replication set costs more — the uploader pays for the availability they're requesting.

Storage fees are collected in the `UploadComplete` transaction. The uploading client's account is debited. If the account can't cover the fee, the transaction is rejected.

### Content purchase fees

When a consumer purchases access to content, they pay the price set by the content owner. This is a direct transfer — the protocol takes no cut at the base layer.

The access policy determines the price. For `public` content, the price is zero. For gated content, the owner sets a price in MOJ via their access policy or Goja script:

```javascript
function evaluate(request) {
  var price = 500000000; // 0.5 MOJ in grains
  if (request.payment < price) {
    return { allowed: false, reason: "insufficient payment" };
  }
  return { allowed: true };
}
```

The `GrantAccess` transaction includes a payment field. If the policy requires payment, the validator verifies the payment amount before granting access. The payment is transferred from the consumer's account to the content owner's account atomically within the transaction.

| Flow | From | To | Amount |
|------|------|----|--------|
| Content purchase | Consumer | Content owner | Price set by access policy |
| Gas fee | Consumer | Fee pool (distributed to validators) | Gas cost of `GrantAccess` tx |

For Casbin-based policies, the price can be stored as a policy attribute and evaluated in the matcher. For Goja policies, the script has full control over pricing logic — tiered pricing, bundles, time-based discounts, whatever the owner wants.

### Fee distribution

Collected fees flow into a per-block fee pool distributed to validators:

| Fee type | Distribution |
|----------|-------------|
| Gas fees | Split proportionally among active validators based on voting power |
| Storage fees | Split among validators in the replication set for that content |
| Content purchase fees | Direct transfer to content owner (not pooled) |

Storage fee distribution is the one exception to pooled distribution — the validators doing the actual storage work get paid directly, not the whole set. This incentivizes validators to accept replication set membership (they earn storage fees) and disincentivizes freeloading.

## Validator rewards

Validators earn from three sources:

1. **Gas fees** — their share of the per-block fee pool.
2. **Storage fees** — their share of upload storage fees for content they replicate.
3. **Block rewards** — newly minted tokens per block (inflationary, governance-controlled).

Block rewards are the primary incentive during early network operation when transaction volume is low. As the network matures and fee revenue grows, governance can taper block rewards to control inflation.

### Staking

Staking is necessary but not sufficient for validator admission — candidates must also pass a social election (see [governance.md](governance.md)). Once admitted, stake determines voting power and reward share. Stake can be slashed for:

- Double-signing (signing conflicting blocks at the same height).
- Extended downtime (missing too many consecutive blocks).
- Provable content leakage (if a validator is caught distributing decrypted content — requires social proof, governed off-chain).

Validators can also be removed via community recall vote, independent of slashing.

Delegators can stake MOJ with a validator and earn a share of that validator's rewards, minus the validator's commission rate.

## Token supply

### Genesis allocation

The initial token supply is minted at genesis and distributed across:

| Allocation | Percentage | Vesting | Purpose |
|-----------|-----------|---------|---------|
| Validator incentive pool | 40% | Linear over 4 years | Block rewards, early validator bootstrapping |
| Core team | 20% | 1 year cliff, 3 year linear | Development, operations |
| Community / ecosystem | 20% | Unlocked at governance discretion | Grants, partnerships, integrations |
| Treasury | 15% | Governance-controlled | Protocol development, emergency fund |
| Faucet / bootstrap | 5% | Immediately available | Initial user onboarding, testnet migration |

### Inflation

Block rewards introduce new tokens at a rate set by governance. The target is a declining inflation schedule:

| Year | Target inflation rate |
|------|---------------------|
| 1 | 8% |
| 2 | 6% |
| 3 | 4% |
| 4+ | 2% (floor) |

Actual rates depend on staking participation — higher staking ratio pushes inflation down (fewer rewards needed to incentivize), lower staking ratio pushes it up (need to attract more stakers). This is the same model Cosmos Hub uses.

## Getting tokens into users' hands

This is the cold-start problem: users need MOJ to pay for gas and content, but MOJ has no value until the network is useful, and the network isn't useful until users have MOJ.

### How other chains solved this

| Chain | Bootstrapping mechanism | Notes |
|-------|------------------------|-------|
| Bitcoin | Mining from genesis | Anyone could mine with a CPU. Tokens had no dollar value for years. Organic growth. |
| Ethereum | Presale (2014) | Sold ETH for BTC before launch. Raised ~$18M. Created an initial holder base. |
| Solana | Private rounds + testnet incentives | Multiple funding rounds, then a public token sale. Testnet participants received tokens. Mainnet launched with an existing community. |
| Cosmos | Fundraiser (2017) | Raised ~$17M in ATOMs. Early validators and delegators bootstrapped the network. |
| Avalanche | Public sale + airdrop | Token sale for initial distribution, airdrops to early testers and community members. |

The common thread: some combination of **pre-sale/fundraise** (creates initial holders with skin in the game), **testnet incentivization** (rewards early participants), and **airdrops/faucets** (reduces friction for new users).

### Mojave bootstrapping

**Phase 1: Testnet with faucet**

A faucet distributes testnet MOJ freely. Anyone can request tokens to experiment with uploads, purchases, and access. This is the development and testing phase — tokens have no monetary value. The faucet is rate-limited per address to prevent abuse.

The faucet is a simple service: a funded account that sends grains to any address that requests them, with a cooldown (e.g. 100 MOJ per address per day). Standard implementation — every Cosmos chain has one.

**Phase 2: Incentivized testnet**

Before mainnet launch, run an incentivized testnet where participation earns mainnet token allocations. Validators who run reliable nodes, artists who upload content, developers who build clients — all earn future mainnet tokens proportional to their contribution. This builds the initial community with people who actually used the system.

**Phase 3: Mainnet launch**

Genesis allocation distributes tokens to:
- Validators from the incentivized testnet (they've proven reliability).
- Early uploaders and consumers from the incentivized testnet.
- The faucet/bootstrap allocation (5%) for continued onboarding.

**Phase 4: Organic circulation**

Once the network is live with real content:
- Artists upload and set prices in MOJ.
- Consumers acquire MOJ (via exchanges, P2P, faucet remnants) to purchase access.
- Validators earn MOJ from fees and block rewards.
- MOJ circulates: consumers → artists → (optionally back to consumers if artists spend on infrastructure, or to exchanges).

### The faucet question

A faucet works for bootstrapping but isn't a long-term solution. For mainnet, users need a real on-ramp:

- **Exchange listings** — once MOJ trades on a DEX or CEX, users can buy it with fiat or other crypto.
- **Bridge from existing chains** — if MOJ is bridged to Solana or Ethereum, users in those ecosystems can swap into it.
- **Fiat on-ramp integration** — partner with a payment processor (MoonPay, Transak, etc.) to let users buy MOJ with a credit card directly in the client.
- **Artists as distributors** — an artist could give away small amounts of MOJ to fans as part of a release campaign. The artist funded their account via an exchange; the fan gets tokens without touching crypto infrastructure.

The protocol doesn't mandate any of these — they're ecosystem-level solutions. The base layer just needs the token to exist and be transferable.

## Transaction fee examples

Concrete numbers to give a feel for the economics (assuming 1 MOJ ≈ $1 USD for illustration — actual value is market-determined):

| Action | Fee (grains) | Fee (MOJ) | ~USD |
|--------|-------------|-----------|------|
| Upload a 40 MB track, 3 images, 10-validator replication | ~2,000,000,000 | ~2.0 | ~$2.00 |
| Publish a release (DDEX metadata) | ~50,000 | ~0.00005 | ~$0.00005 |
| Set access policy | ~50,000 | ~0.00005 | ~$0.00005 |
| Add a Casbin policy rule | ~50,000 | ~0.00005 | ~$0.00005 |
| Grant access (gas only, no content payment) | ~50,000 | ~0.00005 | ~$0.00005 |
| Purchase a track (artist sets price at 1 MOJ) | 1,000,050,000 | ~1.00005 | ~$1.00 |
| Deploy a Goja script (1 KB) | ~500,000 | ~0.0005 | ~$0.0005 |
| Submit an attestation | ~200,000 | ~0.0002 | ~$0.0002 |

Storage is the expensive operation. Everything else is negligible — a user doing normal activity (browsing, purchasing, playing) pays effectively nothing in gas.

## Chain parameters (governance-adjustable)

| Parameter | Default | Controls |
|-----------|---------|----------|
| `min_gas_price` | 100 grains/gas unit | Floor price for gas |
| `storage_rate_per_byte` | 1,000 grains/byte | Storage fee rate |
| `block_reward` | 10 MOJ/block | Inflationary reward per block |
| `inflation_rate` | 8% year 1 | Annual token inflation target |
| `max_script_gas` | 1,000,000 gas units | Gas budget for Goja script execution |
| `min_stake` | 10,000 MOJ | Minimum validator stake |
| `unbonding_period` | 21 days | Time before stake is unlocked after unbonding |
| `slash_double_sign` | 5% of stake | Penalty for equivocation |
| `slash_downtime` | 0.01% of stake | Penalty for extended downtime |
| `faucet_rate_limit` | 100 MOJ/address/day | Testnet faucet distribution rate |
