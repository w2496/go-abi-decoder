package decoder

import (
	"bytes"
	"encoding/hex"
	"log"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ether_address = "0x0000000000000000000000000000000000000000"
)

func ParseABI(input string) abi.ABI {
	contractAbi, err := abi.JSON(strings.NewReader(input))
	if err != nil {
		log.Fatal(err)
	}

	return contractAbi
}

func MergeABIs(abis ...string) abi.ABI {
	mergedABI := abi.ABI{
		Methods: make(map[string]abi.Method),
		Events:  make(map[string]abi.Event),
	}

	for _, jsonStr := range abis {
		contractAbi, err := abi.JSON(bytes.NewReader([]byte(jsonStr)))
		if err != nil {
			log.Fatal("decoder.MergeABIs: error parsing ABI: ", err)
		}

		// Merge Methods
		for name, method := range contractAbi.Methods {
			mergedABI.Methods[name] = method
		}

		// Merge Events
		for name, event := range contractAbi.Events {
			mergedABI.Events[name] = event
		}
	}

	return mergedABI
}

func ToSHA3(data string) string {
	hash := crypto.Keccak256([]byte(data))
	return "0x" + hex.EncodeToString(hash)

}

func IsToken(bytecode string) bool {
	tr := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"[2:]
	return strings.Contains(bytecode, tr)
}

func IsERC721(bytecode string) bool {
	return IsToken(bytecode) && strings.Contains(bytecode, "6352211e")
}

func IsERC20(bytecode string) bool {
	return IsToken(bytecode) && strings.Contains(bytecode, "6352211e")
}

// helper function to detect token standard.
func DetectTokenStandard(bytecode string) string {
	if IsToken(bytecode) && IsERC721(bytecode) {
		return "ERC721"
	}

	// Decimals + ttr
	if IsToken(bytecode) && IsERC20(bytecode) {
		return "ERC20"
	}

	return "UNKNOWN"
}

// detectBytecodes checks if a given bytecode string contains a set of signatures.
//
// This function takes a bytecode string and a list of signatures as input. It
// checks if each signature, after optionally removing the "0x" prefix, can be
// found within the bytecode string. The function ensures that all signatures
// are found in the bytecode without collisions and that the count of found
// signatures matches the total count of signatures.
//
// Parameters:
// - bytecode: The bytecode string to search for signatures.
// - signatures: A list of hex signatures to check for within the bytecode.
//
// Returns:
//   - true if all signatures are found without collisions and the count matches,
//     otherwise false.
//
// Example Usage:
//
//	bytecode := "0x0123456789abcdef"
//	signatures := []string{"0x01", "0x23", "0x45"}
//	result := detectBytecodes(bytecode, signatures)
//	// result will be true if all signatures are found without collisions.
func DetectBytecodes(bytecode string, signatures []string) bool {
	// Sort the signatures by string length
	sort.Slice(signatures, func(i, j int) bool {
		return len(signatures[i]) < len(signatures[j])
	})

	remainingBytecode := bytecode // Make a copy of the original bytecode
	found := 0

	for _, code := range signatures {
		code = strings.TrimPrefix(code, "0x") // Remove "0x" prefix if it exists

		if strings.Contains(remainingBytecode, code) {
			// Remove the found code from the remaining bytecode
			remainingBytecode = strings.Replace(remainingBytecode, code, "", 1)
			found++
		}
	}

	// If all signatures were found without collisions and count matches, return true
	return len(signatures) == found
}
