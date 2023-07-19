package decoder

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/exp/slices"
)

// Storage is a struct that holds all the ABIs and indexed contracts.
type Storage struct {
	AbiList []abi.ABI              // global abi storage that holds all abis from `contracts` folder
	Indexed map[string]*AbiStorage // indexed contracts are basically not thought for this application.
	client  *ethclient.Client
}

// Store is a global variable of type Storage, holding all the ABIs and indexed contracts.
var Store = Storage{
	AbiList: make([]abi.ABI, 0),
	Indexed: make(map[string]*AbiStorage),
}

// IndexedAddresses returns a slice of all the addresses of indexed contracts in Store.
func (store *Storage) IndexedAddresses() []string {
	keys := make([]string, 0)

	for k, _ := range store.Indexed {
		keys = append(keys, k)
	}

	return keys
}

// GetIndexed returns the AbiStorage struct for the given address if it exists in Store.
func (store *Storage) GetIndexed(address string) *AbiStorage {
	if store.Indexed[address] != nil {
		return store.Indexed[address]
	}
	return nil
}

// SetIndexed adds the given abi to the indexed contract with the given address in Store.
func (store *Storage) SetIndexed(address string, input abi.ABI, verified bool, isToken bool, bytecode *string) *AbiStorage {
	store.Indexed[address] = &AbiStorage{
		Address:  common.HexToAddress(address),
		Abi:      input,
		Verified: verified,
		IsToken:  isToken,
		Bytecode: bytecode,
		client:   store.client,
	}

	if bytecode == nil && store.client != nil {
		code, err := store.client.CodeAt(context.Background(), store.Indexed[address].Address, nil)
		if err == nil && code != nil {
			_bytecode := strings.Join([]string{"0x", common.Bytes2Hex(code)}, "")
			store.Indexed[address].Bytecode = &_bytecode
		}
	}

	return store.Indexed[address]
}

// RemoveIndexed removes the indexed contract with the given address from Store.
func (store *Storage) RemoveIndexed(address string) {
	delete(store.Indexed, address)
}

// IsIndexed returns true if the given address exists in Store's indexed contracts.
func (s *Storage) IsIndexed(address string) bool {
	return slices.Contains(s.IndexedAddresses(), address)
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

func (store *Storage) SetClient(client *ethclient.Client) {
	store.client = client
}

func (store *Storage) GetClient() *ethclient.Client {
	return store.client
}

func (store *Storage) RemoveClient() {
	store.client = nil
}
