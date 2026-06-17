package keeper

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds everything a keeper needs. Populate it directly, or via
// LoadConfig to read from environment variables.
type Config struct {
	RpcURL           string
	HorizonURL       string
	Passphrase       string
	KeeperSecret     string
	KeeperName       string
	RegistryContract string
	VaultContract    string
	BlendPool        string
	UsdcAddr         string
	SoroswapRouter   string
	PhoenixRouter    string
	PollInterval     int     // seconds between cycles (3–300)
	MinProfit        float64 // minimum lot/bid ratio to act (> 0)
	SlippageBps      int     // max swap slippage in basis points (0–10000)
}

// LoadConfig reads Config from environment variables with testnet defaults.
// KEEPER_SECRET, REGISTRY_CONTRACT, and VAULT_CONTRACT are required; a missing
// one exits the process with a clear message.
func LoadConfig() Config {
	c := Config{
		RpcURL:           envOr("SOROBAN_RPC", "https://soroban-testnet.stellar.org:443"),
		HorizonURL:       envOr("HORIZON_URL", "https://horizon-testnet.stellar.org"),
		Passphrase:       envOr("NETWORK_PASSPHRASE", "Test SDF Network ; September 2015"),
		KeeperSecret:     mustEnv("KEEPER_SECRET"),
		KeeperName:       envOr("KEEPER_NAME", "nectar-keeper"),
		RegistryContract: mustEnv("REGISTRY_CONTRACT"),
		VaultContract:    mustEnv("VAULT_CONTRACT"),
		BlendPool:        envOr("BLEND_POOL", ""),
		UsdcAddr:         envOr("USDC_CONTRACT", ""),
		SoroswapRouter:   envOr("SOROSWAP_ROUTER", ""),
		PhoenixRouter:    envOr("PHOENIX_ROUTER", ""),
	}
	c.PollInterval = intEnv("POLL_INTERVAL", 10, 3, 300)
	c.MinProfit = floatEnv("MIN_PROFIT", 1.02)
	c.SlippageBps = intEnv("SLIPPAGE_BPS", 100, 0, 10000)
	return c
}

func mustEnv(key string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		fmt.Fprintf(os.Stderr, "keeper-sdk: missing required env %s\n", key)
		os.Exit(1)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, def, min, max int) int {
	v := envOr(key, strconv.Itoa(def))
	n, err := strconv.Atoi(v)
	if err != nil || n < min || n > max {
		fmt.Fprintf(os.Stderr, "keeper-sdk: %s=%q invalid (want integer in [%d,%d])\n", key, v, min, max)
		os.Exit(1)
	}
	return n
}

func floatEnv(key string, def float64) float64 {
	v := envOr(key, strconv.FormatFloat(def, 'f', -1, 64))
	f, err := strconv.ParseFloat(v, 64)
	if err != nil || f <= 0 {
		fmt.Fprintf(os.Stderr, "keeper-sdk: %s=%q invalid (want float > 0)\n", key, v)
		os.Exit(1)
	}
	return f
}
