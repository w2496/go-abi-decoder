package decoder

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var ALL_ABIS = MergeJsonABIs(ABI_GOVERNANCE, ABI_TIMELOCK, ABI_DAO_TOKEN, ABI_UNI_ROUTER)

func TestDecodeMethod(t *testing.T) {
	txHash := common.HexToHash(TX_HASH)
	client, err := ethclient.Dial(PROVIDER)

	if err != nil {
		t.Fatal(err)
	}

	// Create a new instance of the ABI decoder
	decoder := AbiDecoder{}

	// Add the ABI to the decoder
	decoder.SetABI(ALL_ABIS)

	transaction, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Decode a method call
	method := decoder.DecodeMethod(transaction)
	t.Logf("Decoded method: %s", method.Signature)
}

func TestDecodeLogs(t *testing.T) {
	txHash := common.HexToHash(TX_HASH)
	client, err := ethclient.Dial(PROVIDER)

	if err != nil {
		t.Fatal(err)
	}

	// Create a new instance of the ABI decoder
	decoder := AbiDecoder{}

	// Add the ABI to the decoder
	decoder.FromJSON(ABI_DAO_TOKEN)

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Decode an event
	for _, event := range decoder.DecodeLogs(receipt.Logs) {
		t.Logf("Decoded event: %s", event.GetSig())
	}
}

func TestJSONMerge(t *testing.T) {
	abis := MergeJsonABIs(ABI_GOVERNANCE, ABI_TIMELOCK, ABI_DAO_TOKEN, ABI_UNI_ROUTER)

	t.Log("merged json")

	t.Log("methods parsed:", len(abis.Methods))
}

const (
	PROVIDER = "https://rpc-devnet-cardano-evm.c1.milkomeda.com"
	TX_HASH  = "0x4f2316c83db20be4833c81f529f0eb51758ff14e4e455b4cbb203482053477f5"
)
