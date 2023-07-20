package decoder

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// AbiDecoder is a struct used to decode contract ABIs.
type AbiDecoder struct {
	IsVerified      bool              // Indicates whether the contract is verified
	ContractAddress *string           // The contract's address
	Abi             *abi.ABI          // The contract's ABI
	Debug           *bool             // Whether debugging is enabled
	client          *ethclient.Client // The client instance for decoder
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

func (decoder *AbiDecoder) SetClient(client *ethclient.Client) {
	decoder.client = client
}

func (decoder *AbiDecoder) GetClient() *ethclient.Client {
	return decoder.client
}

func (decoder *AbiDecoder) RemoveClient() {
	decoder.client = nil
}

func (decoder *AbiDecoder) Reset() {
	decoder.client = nil
	decoder.Abi = nil
	decoder.ContractAddress = nil
	decoder.Debug = nil
}

func (decoder *AbiDecoder) FilterLogEvents(filter ethereum.FilterQuery) (*ScannedLogs, error) {
	if decoder.client == nil {
		return nil, fmt.Errorf("no provider set for decoder - contract: %v", decoder.ContractAddress)
	}

	logs, err := decoder.client.FilterLogs(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	events := make(ScannedLogs, 0)
	for _, log := range logs {
		decoded := decoder.DecodeLog(&log)

		if decoded != nil {
			events = append(events, *decoded)
		}
	}

	return &events, nil
}

func (decoder *AbiDecoder) DecodeReceipt(transactionHash string) (*ScannedLogs, error) {
	if decoder.client == nil {
		return nil, fmt.Errorf("no provider set for decoder - contract: %v", decoder.ContractAddress)
	}

	receipt, err := decoder.client.TransactionReceipt(context.Background(), common.HexToHash(transactionHash))
	if err != nil {
		return nil, err
	}

	events := make(ScannedLogs, 0)
	if receipt.Logs != nil && len(receipt.Logs) > 0 {
		for _, log := range receipt.Logs {
			decoded := decoder.DecodeLog(log)

			if decoded != nil {
				events = append(events, *decoded)
			}
		}
	}

	return &events, nil
}

func (decoder *AbiDecoder) DecodeTransaction(transactionHash string) (*DecodedMethod, error) {
	if decoder.client == nil {
		return nil, fmt.Errorf("no provider set for decoder - contract: %v", decoder.ContractAddress)
	}

	hash := common.HexToHash(transactionHash)
	transaction, _, err := decoder.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, err
	}

	method := decoder.DecodeMethod(transaction)
	return method, nil
}
