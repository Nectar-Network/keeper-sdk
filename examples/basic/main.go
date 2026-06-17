// Command basic runs a minimal Nectar keeper that fills Blend liquidations.
//
// Configure via environment (KEEPER_SECRET, REGISTRY_CONTRACT, VAULT_CONTRACT,
// BLEND_POOL, USDC_CONTRACT) then:
//
//	go run ./examples/basic
package main

import (
	"log"

	sdk "github.com/Nectar-Network/keeper-sdk"
	"github.com/Nectar-Network/keeper-sdk/adapters/blend"
)

func main() {
	cfg := sdk.LoadConfig()
	k, err := sdk.NewKeeper(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// nil DEX client: seized collateral is returned only when it is already
	// USDC. See the multi-pool example to enable Soroswap/Phoenix conversion.
	k.AddAdapter(blend.NewAdapter(blend.Config{
		PoolAddr:   cfg.BlendPool,
		MinProfit:  cfg.MinProfit,
		HorizonURL: cfg.HorizonURL,
		Passphrase: cfg.Passphrase,
		UsdcAddr:   cfg.UsdcAddr,
	}, nil))

	if err := k.Run(); err != nil {
		log.Fatal(err)
	}
}
