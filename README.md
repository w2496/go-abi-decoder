# Abstract ABI Decoder

This is a simple implementation of an abi-decoder that parses methods and events from a given JSON ABI. Note that this kind of ABI parsing is experimental and can lead to errors. For detailed decoding and indexing of contracts, other micro processes should be used. This module is optimized for predefined contracts.

## Usage

To use the ABI decoder, you can import it into your Go code and use it as follows:

```go
// Import the abi-decoder package
import (
    "github.com/w2496/go-abi-decoder"
)

// Create a new instance of the ABI decoder
decoder := decoder.NewDecoder()

// Set the ABI for the contract you want to decode
abi := `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

// Add the ABI to the decoder
decoder.SetABI(abi)

// Decode a method call
decoded, err := decoder.DecodeMethod("transfer", "0x1234567890123456789012345678901234567890", "100")

// Decode an event
decoded, err := decoder.DecodeEvent("Transfer", "0x1234567890123456789012345678901234567890", "0x0987654321098765432109876543210987654321", "100")
```
