package decoder

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// DecodedLog is a struct for holding decoded Ethereum logs.
type DecodedLog struct {
	Contract  string                 `json:"contract"`  // Contract address of the decoded log.
	Topic     string                 `json:"topic"`     // Event topic hash of the decoded log.
	Signature string                 `json:"signature"` // Event signature of the decoded log.
	Params    map[string]interface{} `json:"params"`    // Parameters of the decoded log.
}

// ToJSONBytes returns the JSON-encoded byte array of the DecodedLog object.
func (data *DecodedLog) ToJSONBytes() []byte {
	b, _ := json.Marshal(data)
	return b
}

// ToJSON returns the JSON-encoded string of the DecodedLog object.
func (data *DecodedLog) ToJSON() string {
	return string(data.ToJSONBytes())
}

// GetParamsJSON returns the JSON-encoded string of the parameters of the DecodedLog object.
func (data *DecodedLog) GetParamsJSON() string {
	if data == nil {
		return "{}"
	}

	b, _ := json.Marshal(data.Params)
	return string(b)
}

// GetSig returns the signature of the DecodedLog object.
func (data *DecodedLog) GetSig() string {
	if data == nil {
		return ""
	}

	return data.Signature
}

// GetSigHash returns the signature hash of the DecodedLog object.
func (data *DecodedLog) GetSigHash() string {
	if data == nil {
		return ""
	}

	return data.Topic
}

// DecodedMethod is a struct for holding decoded Ethereum methods.
type DecodedMethod struct {
	Contract  string                 `json:"contract"`  // Contract address of the decoded method.
	SigHash   string                 `json:"sigHash"`   // Function selector hash of the decoded method.
	Signature string                 `json:"signature"` // Function signature of the decoded method.
	Params    map[string]interface{} `json:"params"`    // Parameters of the decoded method.
}

// ToJSONBytes returns the JSON-encoded byte array of the DecodedMethod object.
func (data *DecodedMethod) ToJSONBytes() []byte {
	b, _ := json.Marshal(data)
	return b
}

// ToJSON returns the JSON-encoded string of the DecodedMethod object.
func (data *DecodedMethod) ToJSON() string {
	return string(data.ToJSONBytes())
}

// GetParamsJSON returns the JSON-encoded string of the parameters of the DecodedMethod object.
func (data *DecodedMethod) GetParamsJSON() string {
	if data == nil {
		return "{}"
	}

	b, _ := json.Marshal(data.Params)
	return string(b)
}

// GetSig returns the signature of the DecodedMethod object.
func (data *DecodedMethod) GetSig() string {
	if data == nil {
		return ""
	}

	return data.Signature
}

// GetSigHash returns the signature hash of the DecodedMethod object.
func (data *DecodedMethod) GetSigHash() string {
	if data == nil {
		return ""
	}

	return "0x" + data.SigHash
}

// AbiStorage is a struct for holding Ethereum ABIs.
type AbiStorage struct {
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

// ToJSONBytes returns the JSON-encoded byte array of the AbiStorage object.
func (data *AbiStorage) ToJSONBytes() []byte {
	b, _ := json.Marshal(data)
	return b
}

// ToJSON returns the JSON-encoded string of the AbiStorage object.
func (data *AbiStorage) ToJSON() string {
	return string(data.ToJSONBytes())
}

// ToJSON returns the JSON-encoded string of the AbiStorage object.
func (data *AbiStorage) GetBytecode(address common.Address) *string {
	if data.Bytecode == nil && data.client != nil {
		code, err := data.client.CodeAt(context.Background(), address, nil)
		if err == nil && code != nil {
			_bytecode := strings.Join([]string{"0x", common.Bytes2Hex(code)}, "")
			*data.Bytecode = _bytecode
		}
	}

	return data.Bytecode
}

// returns a single decoder instance of given AbiStorage object
func (data *AbiStorage) GetDecoder() AbiDecoder {
	contractAddress := data.Address.Hex()
	return AbiDecoder{
		ContractAddress: &contractAddress,
		Abi:             &data.Abi,
		IsVerified:      data.Verified,
	}
}

// gets all signature hashes of given ABI
func (data *AbiStorage) GetSigHashes() []string {
	result := make([]string, 0)

	for _, method := range data.Abi.Methods {
		sigHash := crypto.Keccak256Hash([]byte(method.Sig)).String()[:10]
		result = append(result, sigHash)
	}

	return result
}

// gets all signature hashes of given ABI
func (data *AbiStorage) GetTopics() []string {
	result := make([]string, 0)

	for _, event := range data.Abi.Events {
		topic := crypto.Keccak256Hash([]byte(event.Sig)).String()
		result = append(result, topic)
	}

	return result
}

// gets all signatures of given ABI
func (data *AbiStorage) GetSignatures() []string {
	result := make([]string, 0)

	for _, event := range data.Abi.Events {
		result = append(result, event.Sig)
	}

	for _, method := range data.Abi.Methods {
		result = append(result, method.Sig)
	}

	return result
}

func (data *AbiStorage) ValidateBytecodes() *bool {
	if data.Bytecode == nil {
		return nil
	}
	sigs := make([]string, 0)
	sigs = append(sigs, data.GetSigHashes()...)
	sigs = append(sigs, data.GetTopics()...)
	valid := DetectBytecodes(*data.Bytecode, sigs)
	return &valid
}
