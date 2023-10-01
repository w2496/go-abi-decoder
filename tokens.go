package decoder

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ITknInfo represents the structure of the 'token_info' table.
type ITknInfo struct {
	Address   common.Address
	IsERC20   bool
	IsERC721  bool
	IsERC1155 bool
	Name      string
	Symbol    string
	Decimals  uint8
	Meta      string
}

type ITknStore struct {
	data map[common.Address]*ITknInfo
	abis map[common.Address]*abi.ABI
}

var TknStore = ITknStore{
	data: make(map[common.Address]*ITknInfo),
}

func (store *ITknStore) GetClient() *ethclient.Client {
	return Ctx.eth
}

func (store *ITknStore) SetClient(client *ethclient.Client) {
	SetClient(client)
}

func (store *ITknStore) Connect(nodeUrl string) {
	Connect(nodeUrl)
}

func (store *ITknStore) HasAbi(address common.Address) bool {
	return store.data[address] != nil
}

func (store *ITknStore) SetAbi(tkn common.Address, abis abi.ABI) {
	store.abis[tkn] = &abis
}

func (store *ITknStore) GetAbi(addr common.Address) *abi.ABI {
	var result *abi.ABI
	if store.HasAbi(addr) {
		result = store.abis[addr]
	} else if tkn, err := store.Get(addr); err == nil {
		if tkn.IsERC20 {
			*result = ParseABI(ALL_DEFAULT_ABIS[0])
		} else if tkn.IsERC721 {
			*result = ParseABI(ALL_DEFAULT_ABIS[1])
		} else {
			*result = MergeABIs(ALL_DEFAULT_ABIS[0], ALL_DEFAULT_ABIS[1])
		}
	}

	return result
}

func (store *ITknStore) Has(address common.Address) bool {
	return store.data[address] != nil
}

func (store *ITknStore) Set(nfo *ITknInfo) {
	store.data[nfo.Address] = nfo
}

func (store *ITknStore) Get(address common.Address) (*ITknInfo, error) {
	var result ITknInfo

	if store.Has(address) {
		return store.data[address], nil
	} else {
		// Create a context with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result = queryTokenInfo(ctx, address)
	}

	return &result, nil
}

func (store *ITknStore) BalanceOf(tkn common.Address, addr common.Address) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return getERC20Balance(ctx, addr, tkn)
}

func (store *ITknStore) GetDecoder(contract common.Address) (*AbiDecoder, error) {
	if !store.Has(contract) {
		return nil, fmt.Errorf("can not create decoder, token not in store: %s", contract.Hex())
	}

	info, err := store.Get(contract)
	if err != nil {
		return nil, err
	}

	decoder := info.CreateDecoder()
	return &decoder, err
}

func (tkn *ITknInfo) CreateDecoder() AbiDecoder {
	var contractAddress string

	if tkn.Address.Hex() != EtherAddress {
		contractAddress = tkn.Address.Hex()
	}

	return AbiDecoder{
		ContractAddress: &contractAddress,
		client:          TknStore.GetClient(),
		Abi:             TknStore.GetAbi(tkn.Address),
	}
}

func (tkn *ITknInfo) Query() (*ITknInfo, error) {
	return TknStore.Get(tkn.Address)
}

func (tkn *ITknInfo) GetName() *string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return getName(ctx, tkn.Address)
}

func (tkn *ITknInfo) GetSymbol() *string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return getSymbol(ctx, tkn.Address)
}

func (tkn *ITknInfo) GetDecimals() *uint8 {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return getDecimals(ctx, tkn.Address)
}

func (tkn *ITknInfo) BalanceOf(addr common.Address) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return getERC20Balance(ctx, addr, tkn.Address)
}
