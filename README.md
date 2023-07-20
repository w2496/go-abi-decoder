# Abstract ABI Decoder

This is a simple implementation of an abi-decoder that parses methods and events from a given JSON ABI. Note that this kind of ABI parsing is experimental and can lead to errors. For detailed decoding and indexing of contracts, other micro processes should be used. This module is optimized for predefined contracts.

## Usage

To use the ABI decoder, you can import it into your Go code and use it as follows:

```go
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

```go
package main

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	// Import the abi-decoder package
	kdx "github.com/w2496/go-abi-decoder"
)

func main() {
	client, err := ethclient.Dial("https://rpc-devnet-cardano-evm.c1.milkomeda.com")

	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	abis := kdx.MergeABIs(kdx.ALL_DEFAULT_ABIS...)
	decoder := kdx.AbiDecoder{Abi: &abis}
	decoder.SetClient(client)

	go func() {
		defer wg.Done()
		address := "0xd22861049f6582BcAd6b7a33F211e6fC701DBBBB"
		addresses := make([]common.Address, 0)
		addresses = append(addresses, common.HexToAddress(address))

		events, err := decoder.ScanLogs(ethereum.FilterQuery{
			Addresses: addresses,
		})

		if err != nil {
			panic(err)
		}

		fmt.Println(events.ToJSON())
	}()

	go func() {
		defer wg.Done()
		hash := "0x10ad8530cdad3cf34c765ee728e6cd9cef6bf311bdeb2ed0c7dbe8a32d7a0aa8"
		method, err := decoder.ScanTransaction(hash)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(method.ToJSON()))
	}()

	wg.Wait()
}
```
