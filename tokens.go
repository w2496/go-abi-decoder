package decoder

import (
	"context"
	"fmt"
	"time"

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

func (store *ITknStore) HasTknInfo(address common.Address) bool {
	return store.data[address] != nil
}

func (store *ITknStore) SetTknInfo(nfo *ITknInfo) {
	store.data[nfo.Address] = nfo
}

func (store *ITknStore) GetTknInfo(address common.Address) (*ITknInfo, error) {
	var result ITknInfo

	if store.HasTknInfo(address) {
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
	if !store.HasTknInfo(contract) {
		return nil, fmt.Errorf("can not create decoder, token not in store: %s", contract.Hex())
	}

	info, err := store.GetTknInfo(contract)
	if err != nil {
		return nil, err
	}

	decoder := info.GetDecoder()
	return &decoder, err
}

func (tkn *ITknInfo) GetDecoder() AbiDecoder {
	var contractAddress string

	if tkn.Address.Hex() != EtherAddress {
		contractAddress = tkn.Address.Hex()
	}

	decoder := AbiDecoder{
		ContractAddress: &contractAddress,
		client:          TknStore.GetClient(),
	}

	merged := MergeABIs(ALL_DEFAULT_ABIS[0], ALL_DEFAULT_ABIS[1])

	if tkn.IsERC20 {
		decoder.SetABI(ParseABI(ALL_DEFAULT_ABIS[0]))
	} else if tkn.IsERC721 {
		decoder.SetABI(ParseABI(ALL_DEFAULT_ABIS[1]))
	} else {
		decoder.SetABI(merged)
	}

	return decoder
}

func (tkn *ITknInfo) Query() (*ITknInfo, error) {
	return TknStore.GetTknInfo(tkn.Address)
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
