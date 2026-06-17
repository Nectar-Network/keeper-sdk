# Strategy guide

A keeper's behavior is shaped by a few knobs. Three reference profiles:

## Conservative
Fewer, safer fills; tight slippage; only act when clearly profitable.

```sh
MIN_PROFIT=1.05        # require >= 5% lot/bid margin
SLIPPAGE_BPS=50        # 0.5% max swap slippage
POLL_INTERVAL=15
```
- Higher `MIN_PROFIT` skips marginal auctions (less competition risk, fewer fills).
- Tight `SLIPPAGE_BPS` aborts swaps that would realize a poor DEX price; the
  collateral is held rather than dumped.

## Balanced (default)
```sh
MIN_PROFIT=1.02        # 2% margin
SLIPPAGE_BPS=100       # 1%
POLL_INTERVAL=10
```
Good starting point on testnet — fills most profitable auctions while still
rejecting bad swap prices.

## Aggressive
Maximize fill count; tolerate thinner margins and more slippage; poll fast.

```sh
MIN_PROFIT=1.005       # 0.5% margin
SLIPPAGE_BPS=300       # 3%
POLL_INTERVAL=5
```
- Lower `MIN_PROFIT` competes for more auctions but risks losing races and thin
  profit after fees.
- Wider `SLIPPAGE_BPS` lets more swaps through at worse prices — only sensible in
  deep pools.

## Beyond config: custom adapters
For genuinely different logic (a different protocol, a bespoke profitability
model, multi-hop routing), implement your own `ProtocolAdapter` — see
[adapters.md](adapters.md) and [`examples/custom`](../examples/custom/main.go).
`GetTasks` decides *what* is worth doing (and its priority); `Execute` decides
*how*. The slippage gate in the Blend adapter is oracle-anchored: a swap is
rejected when the DEX quote is worse than the oracle-implied fair value by more
than `SLIPPAGE_BPS`, so a manipulated pool can't bait a bad fill.

## Operational guidance
- Keep the keeper's account funded with XLM (fees) and enough USDC liquidity for
  registry fees.
- Run multiple keepers (different keys/pools) for redundancy; each is stateless.
- Watch logs for `zero returnable proceeds` — it means a fill succeeded but the
  swap didn't, so capital is outstanding (see [troubleshooting.md](troubleshooting.md)).
