# Setup — run a keeper in under 10 minutes

## 1. Prerequisites
- **Go 1.24+**
- The **[Stellar CLI](https://developers.stellar.org/docs/tools/developer-tools/cli/install-cli)** (`stellar`) — for registering + staking on-chain.
- A **funded Stellar testnet account** (your keeper's signing key). Create one and fund it:
  ```sh
  stellar keys generate keeper --network testnet --fund
  stellar keys address keeper      # your keeper's G... address
  stellar keys show keeper         # your keeper's S... secret (KEEPER_SECRET)
  ```

## 2. Nectar testnet addresses (copy-paste)
These are the live contracts to target on **Stellar testnet**. (For mainnet, see the
Nectar `docs/NETWORKS.md`; mainnet launches in Tranche 3.)

| Env var | Value |
|---|---|
| `REGISTRY_CONTRACT` | `CDT257SL2IYDZJIDXEVKI67MYLCKE73JY6WGUTGZOEFXJHG26FJHJDRB` |
| `VAULT_CONTRACT` | `CDZR6VDCPQFOFFKKZ2KMVB67Z54LI5OY73NHBFVI6DR6RE6TL7NN7345` |
| `USDC_CONTRACT` | `CD34YC6FFI2KIE2U4ZPCGQIRPH7UPG5YY2QBYNP25ATSFOQSG73J4VBW` |
| `BLEND_POOL` | `CCEBVDYM32YNYCVNRXQKDFFPISJJCV557CDZEIRBEE4NCV4KHPQ44HGF` |
| `SOROSWAP_ROUTER` (optional) | `CCJUD55AG6W5HAI5LRVNKAE5WDP5XGZBUDS5WNTIVDU7O264UZZE7BRD` |

Registry parameters: **`min_stake` = 100 USDC**, `slash_timeout` = 3600 s, `slash_rate` = 10 %.

## 3. Create a project
```sh
mkdir my-keeper && cd my-keeper
go mod init example.com/my-keeper
go get github.com/Nectar-Network/keeper-sdk@v0.1.2
```

## 4. main.go
Copy [`examples/basic`](../examples/basic/main.go) verbatim — it loads config,
registers the Blend adapter, and runs with graceful shutdown.

## 5. Configure
```sh
export KEEPER_SECRET=$(stellar keys show keeper)
export KEEPER_NAME=my-keeper
export REGISTRY_CONTRACT=CDT257SL2IYDZJIDXEVKI67MYLCKE73JY6WGUTGZOEFXJHG26FJHJDRB
export VAULT_CONTRACT=CDZR6VDCPQFOFFKKZ2KMVB67Z54LI5OY73NHBFVI6DR6RE6TL7NN7345
export USDC_CONTRACT=CD34YC6FFI2KIE2U4ZPCGQIRPH7UPG5YY2QBYNP25ATSFOQSG73J4VBW
export BLEND_POOL=CCEBVDYM32YNYCVNRXQKDFFPISJJCV557CDZEIRBEE4NCV4KHPQ44HGF
export SOROSWAP_ROUTER=CCJUD55AG6W5HAI5LRVNKAE5WDP5XGZBUDS5WNTIVDU7O264UZZE7BRD  # optional: collateral -> USDC
# Defaults are fine for testnet; override if needed:
# export SOROBAN_RPC=https://soroban-testnet.stellar.org
# export NETWORK_PASSPHRASE="Test SDF Network ; September 2015"
```
Every variable is documented in [configuration.md](configuration.md).

## 6. Register + stake on-chain (once)
A keeper must be **registered and staked** in the `KeeperRegistry` before the vault
will let it draw capital. This is a one-time, USDC-spending action, so the SDK never
does it implicitly.

**a. Get testnet USDC to stake** (you need ≥ 100 USDC). On Nectar testnet the USDC is
a mock SAC — request a mint from the Nectar team (Telegram `@kunaldrall`), or on
Circle-USDC deployments use Circle's testnet faucet (faucet.circle.com).

**b. Register** — via the [Nectar keeper page](https://nectarnetwork.fun) (connect
your wallet, one click), or from the CLI:
```sh
KEEPER=$(stellar keys address keeper)
stellar contract invoke --id "$REGISTRY_CONTRACT" --source keeper --network testnet \
  -- register --operator "$KEEPER" --name "my-keeper"
```
The registry pulls your 100 USDC stake. Confirm:
```sh
stellar contract invoke --id "$REGISTRY_CONTRACT" --source keeper --network testnet \
  -- get_keeper --operator "$KEEPER"     # active:true, stake:1000000000
```
(You can also call `k.EnsureRegistered()` once at boot — idempotent, but it stakes
USDC on first call. An unregistered keeper logs a clear warning at startup rather
than failing on its first draw.)

## 7. Run
```sh
go run .
```
You'll see `keeper starting` and, each cycle, either task activity or a quiet loop
when nothing is actionable. Logs are structured (`log/slog`). Ctrl-C / SIGTERM stops
cleanly after the in-flight cycle (the examples use `RunContext` +
`signal.NotifyContext`).

## 8. Deploy (optional)
The keeper is a single static binary — `go build -o keeper .` and run it under any
supervisor (systemd, Docker, Railway, Fly). It is stateless: it reads all state from
chain each cycle and restarts safely. See [operators/docker](../README.md) for a
container setup.

Next: [configuration.md](configuration.md) · [strategies.md](strategies.md) ·
[risk.md](risk.md) · [troubleshooting.md](troubleshooting.md)
