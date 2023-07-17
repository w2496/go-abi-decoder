package decoder

import (
	"bytes"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/exp/slices"
)

// Storage is a struct that holds all the ABIs and indexed contracts.
type Storage struct {
	AbiList []abi.ABI              // global abi storage that holds all abis from `contracts` folder
	Indexed map[string]*AbiStorage // indexed contracts are basically not thought for this application.
}

// Store is a global variable of type Storage, holding all the ABIs and indexed contracts.
var Store = Storage{
	AbiList: make([]abi.ABI, 0),
	Indexed: make(map[string]*AbiStorage),
}

// IndexedAddresses returns a slice of all the addresses of indexed contracts in Store.
func (*Storage) IndexedAddresses() []string {
	keys := make([]string, 0)

	for k, _ := range Store.Indexed {
		keys = append(keys, k)
	}

	return keys
}

// GetIndexed returns the AbiStorage struct for the given address if it exists in Store.
func (*Storage) GetIndexed(address string) *AbiStorage {
	if Store.Indexed[address] != nil {
		return Store.Indexed[address]
	}
	return nil
}

// SetIndexed adds the given abi to the indexed contract with the given address in Store.
func (*Storage) SetIndexed(address string, input abi.ABI, verified bool, isToken bool) *AbiStorage {
	Store.Indexed[address] = &AbiStorage{
		Address:  common.HexToAddress(address),
		Abi:      input,
		Verified: verified,
		IsToken:  isToken,
	}

	return Store.Indexed[address]
}

// RemoveIndexed removes the indexed contract with the given address from Store.
func (data *Storage) RemoveIndexed(address string) {
	delete(Store.Indexed, address)
}

// IsIndexed returns true if the given address exists in Store's indexed contracts.
func (store *Storage) IsIndexed(address string) bool {
	return slices.Contains(Store.IndexedAddresses(), address)
}

// DecodeLogs decodes an array of Ethereum logs into an array of DecodedLogs using the DecodeLog function.
// Any nil DecodedLogs are not included in the result array.
func (store *Storage) DecodeLogs(vLogs []*types.Log) []*DecodedLog {
	var decodedLogs []*DecodedLog

	for _, log := range vLogs {
		decoded := store.DecodeLog(log)
		if decoded != nil {
			decodedLogs = append(decodedLogs, decoded)
		}
	}

	return decodedLogs
}

// DecodeLog decodes a single Ethereum log entry and returns a `DecodedLog` object that contains
// the decoded values. This function checks if the log entry corresponds to a token transfer event
// and if so, it determines whether it is an ERC20 or ERC721 transfer and picks the right ABI for
// decoding the log data. If the log cannot be decoded or is not a token transfer event, it returns
// nil. This function iterates through all ABIs from Store.AbiList to attempt to decode the log
// data using each ABI in turn. If the log can be decoded by any ABI, it returns a `DecodedLog`
// object containing the decoded values. Otherwise, it returns nil.
func (store *Storage) DecodeLog(vLog *types.Log) *DecodedLog {
	// Cache frequently-used variables to avoid overhead on every call to DecodeLog.
	abis := store.AbiList

	// Check all other ABIs.
	for _, contractAbi := range abis {
		abiDecoder := AbiDecoder{Abi: &contractAbi}
		decoded := abiDecoder.DecodeLog(vLog)
		if decoded != nil && decoded.Signature != "" {
			return decoded
		}
	}

	return nil
}

// DecodeMethod decodes a single Ethereum transaction and returns a `DecodedMethod` object that
// contains the decoded function signature and arguments. This function iterates through all ABIs
// from `Store.AbiList` to attempt to decode the transaction using each ABI in turn. If the
// transaction can be decoded by any ABI, it returns a `DecodedMethod` object containing the
// decoded function signature and arguments. Otherwise, it returns nil.
func (store *Storage) DecodeMethod(tx *types.Transaction) *DecodedMethod {
	for _, contractAbi := range store.AbiList {
		abiDecoder := AbiDecoder{Abi: &contractAbi}
		decoded := abiDecoder.DecodeMethod(tx)
		if decoded != nil {
			return decoded
		}
	}

	return nil
}

func (store *Storage) ParseAndAddABIs(abis ...string) {
	for _, abi := range abis {
		store.AbiList = append(store.AbiList, ParseABI(abi))
	}
}

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
			log.Fatal("error parsing ABI: ", err)
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
