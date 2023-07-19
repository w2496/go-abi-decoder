package decoder

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type Params map[string]interface{}

func (m *Params) MarshalJSON() ([]byte, error) {
	regex := regexp.MustCompile(`^"0x[0-9a-fA-F]{40}"$`)
	result := "{"
	var parts []string

	for k, v := range *m {
		part, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		if regex.MatchString(string(part)) {
			addr := common.HexToAddress(strings.ReplaceAll(string(part), "\"", ""))
			parts = append(parts, fmt.Sprintf(`"%s":"%s"`, k, addr.Hex()))
		} else {
			parts = append(parts, fmt.Sprintf(`"%s":%s`, k, string(part)))
		}
	}

	result += strings.Join(parts, ",")
	result += "}"

	return []byte(result), nil
}

type ScannedLogs []DecodedLog

func (l *ScannedLogs) ToJSONBytes() []byte {
	res, err := json.Marshal(l)
	if err != nil {
		panic(err)
	}

	return res
}

func (l *ScannedLogs) ToJSON() string {
	return string(l.ToJSONBytes())
}

// DecodedLog is a struct for holding decoded Ethereum logs.
type DecodedLog struct {
	Contract        string `json:"contract"`        // Contract address of the decoded log.
	Topic           string `json:"topic"`           // Event topic hash of the decoded log.
	Signature       string `json:"signature"`       // Event signature of the decoded log.
	Params          Params `json:"params"`          // Parameters of the decoded log.
	TransactionHash string `json:"transactionHash"` // Transaction hash of the decoded log.
	LogIndex        uint   `json:"logIndex"`        // Index of the decoded log
	BlockNumber     uint64 `json:"blockNumber"`     // blockNumber of given decoded log
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
	TransactionHash string `json:"transactionHash"` // Transaction hash of the decoded method.
	Contract        string `json:"contract"`        // Contract address of the decoded method.
	SigHash         string `json:"sigHash"`         // Function selector hash of the decoded method.
	Signature       string `json:"signature"`       // Function signature of the decoded method.
	Params          Params `json:"params"`          // Parameters of the decoded method.
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
