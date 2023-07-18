package decoder

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/exp/maps"
)

var (
	target_provider = "https://rpc-devnet-cardano-evm.c1.milkomeda.com"
	target_tx_hash  = "0x4f2316c83db20be4833c81f529f0eb51758ff14e4e455b4cbb203482053477f5"
	target_contract = "0xd22861049f6582BcAd6b7a33F211e6fC701DBBBB"
	all_abis_parsed = MergeABIs(ALL_DEFAULT_ABIS...)
)

var user_abi = `
[
    {
    "anonymous": false,
    "inputs": [
      {
        "indexed": true,
        "internalType": "address",
        "name": "delegator",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "fromDelegate",
        "type": "address"
      },
      {
        "indexed": true,
        "internalType": "address",
        "name": "toDelegate",
        "type": "address"
      }
    ],
    "name": "DelegateChanged",
    "type": "event"
  },
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "delegatee",
        "type": "address"
      }
    ],
    "name": "delegate",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  }
]
`

func init() {
	fmt.Println("initializing provider", target_provider)
	client, err := ethclient.Dial(target_provider)

	if err != nil {
		log.Fatal(err)
	}
	Store.client = client
}

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

	// Create a new instance of the ABI decoder
	decoder := AbiDecoder{}

	// // Add the ABI to the decoder
	decoder.SetABI(ParseABI(user_abi))

	transaction, _, err := Store.client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Decode a method call
	method := decoder.DecodeMethod(transaction)
	t.Logf("Decoded method: %s - sigHash: %s", method.Signature, method.SigHash)

	method = Store.DecodeMethod(transaction)
	t.Logf("Decoded method (Store): %s - sigHash: %s", method.Signature, method.SigHash)

	t.Logf(method.ToJSON())
}

func TestDecodeLogs(t *testing.T) {
	txHash := common.HexToHash(target_tx_hash)

	// Create a new instance of the ABI decoder
	decoder := AbiDecoder{
		Abi: &all_abis_parsed,
	}

	// Add the ABI to the decoder
	// decoder.FromJSON(abi_dao_token)

	receipt, err := Store.client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		t.Fatal(err)
	}

	// Decode an event
	for _, event := range decoder.DecodeLogs(receipt.Logs) {
		t.Logf("Decoded event: %s - sigHash: %s - contract: %s", event.GetSig(), event.Topic, event.Contract)
		t.Log(event.GetParamsJSON())
	}

	// Decode an event using Store Decoder
	for _, event := range Store.DecodeLogs(receipt.Logs) {
		t.Logf("Decoded event (Store): %s - topic: %s - contract: %s", event.GetSig(), event.Topic, event.Contract)
		t.Log(event.GetParamsJSON())
	}
}

func TestIndexedDecoder(t *testing.T) {
	Store.SetIndexed(target_contract, ParseABI(ALL_DEFAULT_ABIS[12]), true, false, nil)
	s := Store.GetIndexed(target_contract)
	t.Logf(`bytecode: %v - size: %v`, s.Bytecode, len(*s.Bytecode))

	if s.Address.Hex() != target_contract {
		t.Fatalf("invalid result address: %s", s.Address)
	}

	decoder := s.GetDecoder()

	maps.Keys(decoder.Abi.Methods)

	t.Logf("Decoder contains %v methods and %v events", len(decoder.Abi.Methods), len(decoder.Abi.Events))
	t.Logf("Decoder assigned to contract %v", *decoder.ContractAddress)

	for signature, sigHash := range s.GetSigHashes() {
		t.Logf("signature: %v - sigHash: %v", signature, sigHash)
	}

	for _, topic := range s.GetTopics() {
		t.Logf("topic: %v", topic)
	}

	for _, signature := range s.GetSignatures() {
		t.Logf("signature: %v", signature)
	}

	t.Logf("isToken: %v", IsToken(*s.Bytecode))
	t.Logf("isERC20: %v", IsERC20(*s.Bytecode))
	t.Logf("isERC721: %v", IsERC721(*s.Bytecode))

	t.Logf("topics in bytecode: %v", DetectBytecodes(*s.Bytecode, s.GetSigHashes()))
}

func TestSHA3(t *testing.T) {
	hash := ToSHA3("DelegateChanged(address,address,address)")
	if hash != "0x3134e8a2e6d97e929a7e54011ea5485d7d196dd5f0ba4d4ef95803e8e3fc257f" {
		t.Fatalf("solidity returned incorrect hash: %v", hash)
	}
}
