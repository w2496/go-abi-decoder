package decoder

import (
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

// AbiDecoder is a struct used to decode contract ABIs.
type AbiDecoder struct {
	IsVerified      bool     // Indicates whether the contract is verified
	ContractAddress *string  // The contract's address
	Abi             *abi.ABI // The contract's ABI
	Debug           *bool    // Whether debugging is enabled
}

// checkAbi checks if the ABI has been loaded into the decoder instance.
// If not, it throws a fatal error.
func checkAbi(decoder *AbiDecoder) {
	if decoder.Abi == nil {
		log.Fatal("NO ABI LOADED TO INSTANCE", decoder)
	}
}

// SetABI sets the contract ABI in the decoder instance and returns it.
func (decoder *AbiDecoder) SetABI(contractAbi abi.ABI) abi.ABI {
	decoder.Abi = &contractAbi
	return *decoder.Abi
}

// FromJSON decodes the ABI from JSON and sets it in the decoder instance.
// It returns the contract ABI.
func (decoder *AbiDecoder) FromJSON(abis string) abi.ABI {
	contractAbi, err := abi.JSON(strings.NewReader(abis))
	if err != nil {
		log.Fatal(err)
	}

	decoder.Abi = &contractAbi
	return *decoder.Abi
}

func (s *AbiDecoder) MergeAddABIs(abis ...string) abi.ABI {
	*s.Abi = MergeABIs(abis...)
	return *s.Abi
}

// DecodeLog decodes the log and returns the decoded log.
// It checks if the ABI has been loaded in the decoder instance.
func (decoder *AbiDecoder) DecodeLog(vLog *types.Log) *DecodedLog {
	checkAbi(decoder)
	return parseLog(vLog, *decoder.Abi, decoder.Debug)
}

// DecodeLogs decodes a slice of Ethereum logs using the ABI specified in the `AbiDecoder`. It
// returns a slice of `DecodedLog` objects that contain the decoded event signature and arguments
// for each log. The function first checks that an ABI has been specified using the `checkAbi()`
// helper function, and then iterates through each log, calling the `parseLog()` function to
// attempt to decode the log using the specified ABI. If the log can be decoded, a `DecodedLog`
// object is added to the result slice. Finally, the function returns the result slice.
func (decoder *AbiDecoder) DecodeLogs(vLogs []*types.Log) []*DecodedLog {
	checkAbi(decoder)
	result := make([]*DecodedLog, 0, len(vLogs))

	for _, v := range vLogs {
		if decoded := parseLog(v, *decoder.Abi, decoder.Debug); decoded != nil {
			result = append(result, decoded)
		}
	}

	return result
}

// DecodeMethod decodes the method of a given transaction using the ABI loaded in the decoder.
// It takes a types.Transaction as an input and returns a pointer to a DecodedMethod if the
// method was successfully decoded, or nil if not.
func (decoder *AbiDecoder) DecodeMethod(tx *types.Transaction) *DecodedMethod {
	// Check if the ABI has been loaded
	checkAbi(decoder)

	// Parse the method
	return parseMethod(tx, *decoder.Abi, decoder.Debug)
}
