package decoder

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// IndexedABI is a struct for holding Ethereum ABIs.
type IndexedABI struct {
	Address  common.Address `json:"address"`            // Address of the contract the ABI belongs to.
	Abi      abi.ABI        `json:"abi"`                // The ABI of the contract.
	Bytecode *string        `json:"bytecode,omitempty"` // Bytecode of the contract the ABI belongs to.
	IsToken  bool           `json:"isToken"`            // Current ABI is a Token
	Verified bool           `json:"verified"`           // Whether the ABI has been verified.
	IsERC721 *bool          `json:"isERC721,omitempty"` // contract is NFT Token
	Name     *string        `json:"name,omitempty"`     // Name of the contract
	Pragma   *string        `json:"pragma,omitempty"`   // Pragma Solidity Version of contract
	Source   *string        `json:"source,omitempty"`   // Solidity source code of contract
	client   *ethclient.Client
}

// ToJSONBytes returns the JSON-encoded byte array of the IndexedABI object.
func (data *IndexedABI) ToJSONBytes() []byte {
	b, _ := json.Marshal(data)
	return b
}

// ToJSON returns the JSON-encoded string of the IndexedABI object.
func (data *IndexedABI) ToJSON() string {
	return string(data.ToJSONBytes())
}

// ToJSON returns the JSON-encoded string of the IndexedABI object.
func (data *IndexedABI) GetBytecode() *string {
	if data.Bytecode == nil && data.client != nil {
		data.Bytecode = getBytecode(data.client, data.Address)
	}

	return data.Bytecode
}

// returns a single decoder instance of given IndexedABI object
func (data *IndexedABI) GetDecoder() AbiDecoder {
	contractAddress := data.Address.Hex()

	if data.Bytecode == nil {
		data.Bytecode = data.GetBytecode()
		fmt.Println("bytecode loaded")
	}

	return AbiDecoder{
		ContractAddress: &contractAddress,
		Abi:             &data.Abi,
		IsVerified:      data.Verified,
		client:          data.client,
	}
}

// gets all signature hashes of given IndexedABI
func (data *IndexedABI) GetSigHashes() []string {
	result := make([]string, 0)

	for _, method := range data.Abi.Methods {
		sigHash := ToSHA3(method.Sig)
		result = append(result, sigHash[:10])
	}

	sort.Slice(result, func(i, j int) bool {
		return len(result[i]) < len(result[j])
	})

	return result
}

// gets all signature hashes of given IndexedABI
func (data *IndexedABI) GetTopics() []string {
	result := make([]string, 0)

	for _, event := range data.Abi.Events {
		topic := ToSHA3(event.Sig)
		result = append(result, topic)
	}

	sort.Slice(result, func(i, j int) bool {
		return len(result[i]) < len(result[j])
	})

	return result
}

// gets all signatures of given IndexedABI
func (data *IndexedABI) GetSignatures() []string {
	result := make([]string, 0)

	for _, event := range data.Abi.Events {
		result = append(result, event.Sig)
	}

	for _, method := range data.Abi.Methods {
		result = append(result, method.Sig)
	}

	sort.Slice(result, func(i, j int) bool {
		return len(result[i]) < len(result[j])
	})

	return result
}

func (data *IndexedABI) ValidateBytecodes() *bool {
	if data.Bytecode == nil {
		return nil
	}
	sigs := make([]string, 0)
	sigs = append(sigs, data.GetSigHashes()...)
	sigs = append(sigs, data.GetTopics()...)
	valid := DetectBytecodes(*data.Bytecode, sigs)
	return &valid
}

func (indexed *IndexedABI) SetClient(client *ethclient.Client) {
	indexed.client = client
}

func (indexed *IndexedABI) GetClient() *ethclient.Client {
	return indexed.client
}

func (indexed *IndexedABI) RemoveClient() {
	indexed.client = nil
}
