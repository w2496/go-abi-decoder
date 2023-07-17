# Abstract ABI Decoder

This is a simple implementation of an abi-decoder that parses methods and events from a given JSON ABI. Note that this kind of ABI parsing is experimental and can lead to errors. For detailed decoding and indexing of contracts, other micro processes should be used. This module is optimized for predefined contracts.

## Usage

To use the ABI decoder, you can import it into your Go code and use it as follows:

```golang
package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	// Import the abi-decoder package
	kdx "github.com/w2496/go-abi-decoder"
)

func main() {
	txHash := common.HexToHash("0x4f2316c83db20be4833c81f529f0eb51758ff14e4e455b4cbb203482053477f5")
	client, err := ethclient.Dial("https://rpc-devnet-cardano-evm.c1.milkomeda.com")

	if err != nil {
		panic(err)
	}

	abi := kdx.ParseABI(kdx.ALL_DEFAULT_ABIS[12])

	// Create a new instance of the ABI decoder
	decoder := kdx.AbiDecoder{
		Abi: &abi,
	}

	transaction, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		panic(err)
	}

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		panic(err)
	}

	// Decode a method call
	method := decoder.DecodeMethod(transaction)
	fmt.Println(method.ToJSON())

	// Decode an event
	for _, event := range decoder.DecodeLogs(receipt.Logs) {
		fmt.Println(event.ToJSON())
	}
}
```
