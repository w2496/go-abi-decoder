package decoder

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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
func (*Storage) SetIndexed(address string, input abi.ABI, verified bool) *AbiStorage {
	Store.Indexed[address] = &AbiStorage{
		Address:  common.HexToAddress(address),
		Abi:      input,
		Verified: verified,
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
