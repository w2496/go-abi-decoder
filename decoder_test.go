package decoder

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	target_provider = "https://rpc-devnet-cardano-evm.c1.milkomeda.com"
	target_tx_hash  = "0x4f2316c83db20be4833c81f529f0eb51758ff14e4e455b4cbb203482053477f5"
	target_contract = "0xd22861049f6582BcAd6b7a33F211e6fC701DBBBB"
	all_abis_parsed = MergeABIs(ALL_DEFAULT_ABIS...)
)

func TestMergeAbISs(t *testing.T) {
	abis := MergeABIs(ALL_DEFAULT_ABIS...)

	t.Log("merged json")
	t.Log("methods parsed:", len(abis.Methods))
}

func TestAbiStore(t *testing.T) {
	Store.ParseAndAddABIs(ALL_DEFAULT_ABIS...)
	t.Logf("%v ABIS added to Store", len(Store.AbiList))
}

func TestDecodeMethod(t *testing.T) {
	txHash := common.HexToHash(target_tx_hash)
	client, err := ethclient.Dial(target_provider)

	if err != nil {
		t.Fatal(err)
	}

	// Create a new instance of the ABI decoder
	decoder := AbiDecoder{}

	// Add the ABI to the decoder
	decoder.SetABI(all_abis_parsed)

	transaction, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Decode a method call
	method := decoder.DecodeMethod(transaction)
	t.Logf("Decoded method: %s", method.Signature)

	method = Store.DecodeMethod(transaction)
	t.Logf("Decoded method (Store): %s", method.Signature)
}

func TestDecodeLogs(t *testing.T) {
	txHash := common.HexToHash(target_tx_hash)
	client, err := ethclient.Dial(target_provider)

	if err != nil {
		t.Fatal(err)
	}

	// Create a new instance of the ABI decoder
	decoder := AbiDecoder{}

	// Add the ABI to the decoder
	decoder.FromJSON(abi_dao_token)

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Decode an event
	for _, event := range decoder.DecodeLogs(receipt.Logs) {
		t.Logf("Decoded event: %s - contract: %s", event.GetSig(), event.Contract)
	}

	// Decode an event using Store Decoder
	for _, event := range Store.DecodeLogs(receipt.Logs) {
		t.Logf("Decoded event (Store): %s - contract: %s", event.GetSig(), event.Contract)
	}
}

func TestIndexedDecoder(t *testing.T) {
	Store.SetIndexed(target_contract, ParseABI(ALL_DEFAULT_ABIS[11]), true, false)
	s := Store.GetIndexed(target_contract)

	if s.Address.Hex() != target_contract {
		t.Fatalf("invalid result address: %s", s.Address)
	}
}
