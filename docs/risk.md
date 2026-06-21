# Risk Management

Running a keeper means putting capital and a staked bond to work autonomously.
This guide consolidates every way a keeper operator can lose value and how the
SDK defends against each. Read it before pointing a keeper at a vault you care
about.

## 1. Stake and slashing

To register, an operator stakes USDC into the `KeeperRegistry`. That stake is
**slashable**: if the keeper draws vault capital and fails to return proceeds
within the configured `slash_timeout` (testnet default: 3,600 s), a portion of
the stake (`slash_rate_bps`, testnet default: 1,000 = 10%) is transferred to the
vault as compensation.

Mitigations:
- The keeper reconciles stale draws automatically on the next cycle (it returns
  whatever USDC actually arrived) rather than leaving a draw open until timeout.
- Keep liquid USDC/XLM in the keeper account so a return transaction never fails
  for lack of fees while the slash clock is running.
- Size `MIN_PROFIT` so you are not drawing capital for marginal fills that can
  flip to a loss after slippage.

## 2. Slippage and oracle anchoring

After a fill the keeper sells seized collateral for USDC. Two independent
defenses bound the price you accept:

- **On-chain `amount_out_min`** — every swap passes a minimum-out derived from the
  router quote and `SLIPPAGE_BPS` (default **100 bps = 1%**). The swap reverts
  on-chain if execution would breach it. `minOutForSlippage` uses big-integer
  math so a manipulated/garbage quote cannot overflow into a tiny floor.
- **Pre-trade oracle floor** — when an oracle price is available, the keeper
  refuses to broadcast at all if the router quote is below `oracle × (1 −
  slippage)`, and it will not fall back to another venue at that price either.

Caveat: the oracle floor is best-effort. On pools the oracle cannot price it
returns 0 and the floor is disabled — only the quote-anchored `amount_out_min`
applies. Treat unpriced collateral as higher risk and tighten `SLIPPAGE_BPS`.

A programmatically-built `Config{}` now defaults `SlippageBps` to 100; set it
explicitly if you want a different bound.

## 3. Post-send ambiguity (no double-execution)

A swap can be broadcast but its final status unknown (RPC timeout / reorg
window). Re-selling the same collateral on a second venue would dump it twice.
The SDK classifies this case (`soroban.IsTxStatusUnknown`) and **stops** — it
never falls back to Phoenix after a sent-but-unknown Soroswap swap. The
stale-draw reconciliation on the next cycle then books whatever USDC arrived.
Swaps are never blindly retried for the same reason.

## 4. Capital-outstanding exposure

Between `Draw` and `ReturnProceeds` the keeper holds vault capital. If the
process dies mid-cycle the draw is recorded on-chain and recovered next cycle —
but until then that capital is outstanding and the slash clock is ticking. Run
one keeper instance per registered key; two instances sharing a key will fight
over sequence numbers and lose races, increasing the odds of a failed return.

## 5. Keys and funding

- Use a **dedicated keypair per keeper instance**. Never reuse a key that holds
  personal funds — the keeper account is exposed to slashing and to any bug in a
  third-party adapter.
- Store `KEEPER_SECRET` as a platform secret (e.g. Railway "mark as secret"),
  never in shell history or a committed file.
- Keep enough XLM for transaction fees and enough liquid USDC to cover the stake
  plus a working buffer.

## 6. Third-party adapters

The keeper loop isolates each adapter: a panic in one adapter is recovered and
logged, so a faulty `ProtocolAdapter` degrades to a skipped cycle instead of
crashing the keeper. Still, an adapter you did not write can submit transactions
with your key — review adapter source before registering it.

## See also

- [Configuration](./configuration.md) — every environment variable, including
  `SLIPPAGE_BPS`, `MIN_PROFIT`, and the slash-related registry parameters.
- [Strategies](./strategies.md) — conservative / balanced / aggressive profit
  thresholds.
- [Troubleshooting](./troubleshooting.md) — diagnosing failed returns and swaps.
