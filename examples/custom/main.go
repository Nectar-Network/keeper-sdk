// Command custom shows how to implement a ProtocolAdapter with a bespoke
// strategy. This stub discovers no tasks; replace GetTasks/Execute with your own
// protocol logic. Run with the standard env from LoadConfig:
//
//	go run ./examples/custom
package main

import (
	"log"

	"github.com/stellar/go/keypair"

	sdk "github.com/Nectar-Network/keeper-sdk"
	"github.com/Nectar-Network/keeper-sdk/soroban"
)

// myAdapter is a minimal ProtocolAdapter implementation.
type myAdapter struct{}

func (myAdapter) Name() string { return "my-protocol" }

// GetTasks scans your protocol and returns actionable tasks (reads only).
func (myAdapter) GetTasks(rpc *soroban.Client) ([]sdk.Task, error) {
	// e.g. rpc.SimulateRead(...) to find work, then return []sdk.Task{...}.
	return nil, nil
}

// Execute performs one task, drawing/returning capital via vault when needed.
func (myAdapter) Execute(rpc *soroban.Client, kp *keypair.Full, task sdk.Task, vault sdk.VaultClient) (*sdk.Result, error) {
	// e.g. vault.Draw(amount); submit a tx via rpc.Invoke(...); vault.ReturnProceeds(got, ms).
	return &sdk.Result{Success: true}, nil
}

// EstimateCapital returns the USDC a task needs (0 if it uses no vault capital).
func (myAdapter) EstimateCapital(task sdk.Task) (int64, error) { return 0, nil }

// Compile-time check that myAdapter satisfies the interface.
var _ sdk.ProtocolAdapter = myAdapter{}

func main() {
	cfg := sdk.LoadConfig()
	k, err := sdk.NewKeeper(cfg)
	if err != nil {
		log.Fatal(err)
	}
	k.AddAdapter(myAdapter{})
	if err := k.Run(); err != nil {
		log.Fatal(err)
	}
}
