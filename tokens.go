package decoder

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
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
	NodeUrl *string
	EIP1559 *bool
	client  *ethclient.Client
	data    map[common.Address]*ITknInfo
}

var TknStore = ITknStore{
	data:    make(map[common.Address]*ITknInfo),
	EIP1559: nil,
	NodeUrl: nil,
}

func init() {
	if TknStore.client == nil && TknStore.client != nil {
		TknStore.SetClient(Store.client)
	}
}

func (store *ITknStore) GetClient() *ethclient.Client {
	return TknStore.client
}

func (store *ITknStore) SetClient(client *ethclient.Client) {
	TknStore.client = client

	i, err := IsEIP1559(client, context.Background())

	if err == nil && i != nil {
		TknStore.EIP1559 = i
	}
}

func (store *ITknStore) Connect(nodeUrl string) {
	store.NodeUrl = &nodeUrl
	client, err := ethclient.Dial(*store.NodeUrl)

	if err != nil {
		panic(err)
	}

	store.SetClient(client)
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
		client, ctx, cancel := requireClientContext()
		defer cancel()
		result = queryTokenInfo(client, ctx, address)
	}

	return &result, nil
}

func (store *ITknStore) BalanceOf(tkn common.Address, addr common.Address) (uint64, error) {
	client, ctx, cancel := requireClientContext()
	defer cancel()
	return getERC20Balance(client, ctx, addr, tkn)
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
	client, ctx, cancel := requireClientContext()
	defer cancel()
	return getName(client, ctx, tkn.Address)
}

func (tkn *ITknInfo) GetSymbol() *string {
	client, ctx, cancel := requireClientContext()
	defer cancel()
	return getSymbol(client, ctx, tkn.Address)
}

func (tkn *ITknInfo) GetDecimals() *uint8 {
	client, ctx, cancel := requireClientContext()
	defer cancel()
	return getDecimals(client, ctx, tkn.Address)
}

func (tkn *ITknInfo) BalanceOf(addr common.Address) (uint64, error) {
	if TknStore.client == nil {
		return 0, fmt.Errorf("no client connected to token store")
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return getERC20Balance(TknStore.client, ctx, addr, tkn.Address)
}

func getSymbol(client *ethclient.Client, ctx context.Context, contract common.Address) *string {
	msg := ethereum.CallMsg{
		To: &contract, Data: common.Hex2Bytes("95d89b41"),
	}
	symbol, err := client.CallContract(ctx, msg, nil)

	if err != nil {
		return nil
	}

	result := ToAscii(symbol)

	return &result
}

func getName(client *ethclient.Client, ctx context.Context, contract common.Address) *string {
	msg := ethereum.CallMsg{
		To: &contract, Data: common.Hex2Bytes("06fdde03"),
	}
	name, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil
	}

	out0 := ToAscii(name)

	return &out0
}

func getDecimals(client *ethclient.Client, ctx context.Context, contract common.Address) *uint8 {
	msg := ethereum.CallMsg{
		To: &contract, Data: common.Hex2Bytes("313ce567"),
	}
	decimals, err := client.CallContract(ctx, msg, nil)

	if err != nil {
		return nil
	}

	result := uint8(common.BytesToHash(decimals).Big().Uint64())
	return &result
}

func getERC20Balance(client *ethclient.Client, ctx context.Context, address common.Address, contractAddress common.Address) (uint64, error) {
	// Create an instance of the ERC-20 contract ABI
	contractAbi, err := abi.JSON(strings.NewReader(ALL_DEFAULT_ABIS[0]))
	if err != nil {
		return 0, err
	}

	// Build a call data to get the balance of the address
	data, err := contractAbi.Pack("balanceOf", address)
	if err != nil {
		return 0, err
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	// Perform the call to the ERC-20 contract
	result, err := Store.client.CallContract(ctx, msg, nil)
	if err != nil {
		return 0, err
	}

	// Unpack the result to get the balance as a big.Int
	var balance *big.Int
	err = contractAbi.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return 0, err
	}

	return balance.Uint64(), nil
}

func queryTokenInfo(client *ethclient.Client, ctx context.Context, address common.Address, bytecodes ...string) ITknInfo {
	var code *string
	if len(bytecodes) > 0 {
		var byteSlice []string
		for _, bc := range bytecodes {
			byteSlice = append(byteSlice, strings.TrimPrefix(bc, "0x"))
		}

		joined := "0x" + strings.TrimPrefix(strings.Join(byteSlice, ""), "0x")
		code = &joined

	} else {
		code = getBytecode(Store.client, address)
	}

	symbol := getSymbol(client, ctx, address)
	name := getName(client, ctx, address)
	decimals := getDecimals(client, ctx, address)

	isErc20 := IsERC20(*code)
	isErc721 := IsERC721(*code)
	isErc1155 := isErc20 && isErc721

	result := ITknInfo{
		Address:   address,
		IsERC20:   isErc20,
		IsERC721:  isErc721,
		IsERC1155: isErc1155,
		Name:      *name,
		Symbol:    *symbol,
		Decimals:  *decimals,
		Meta:      "{}",
	}
	return result
}

func requireClientContext() (*ethclient.Client, context.Context, context.CancelFunc) {
	client := TknStore.client
	if TknStore.client == nil && TknStore.NodeUrl == nil && Store.client == nil {
		panic(fmt.Errorf("no client connected to token store"))
	}

	if TknStore.client == nil && Store.client != nil {
		client = Store.client
	}

	if client == nil && TknStore.NodeUrl != nil {
		TknStore.Connect(*TknStore.NodeUrl)
		client = TknStore.client
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	return client, ctx, cancel
}
