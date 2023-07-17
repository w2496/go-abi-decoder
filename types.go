package decoder

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
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
	IsToken  bool           `json:"isToken"`  // Current ABI is a Token
	Verified bool           `json:"verified"` // Whether the ABI has been verified.
	Address  common.Address `json:"address"`  // Address of the contract the ABI belongs to.
	Abi      abi.ABI        `json:"abi"`      // The ABI.
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
