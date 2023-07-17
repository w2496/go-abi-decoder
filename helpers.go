package decoder

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/exp/slices"
)

func ToSHA3(data string) string {
	hash := crypto.Keccak256([]byte(data))
	return "0x" + hex.EncodeToString(hash)

}

// parseMethod extracts the method signature and its parameters from the input data of a transaction, using
// the provided contract ABI to decode the input data. It returns a DecodedMethod object containing the contract
// address, method signature, signature hash, and the decoded method parameters as a map[string]interface{}.
// If there is an error while decoding the input data or the method signature is not found in the ABI, it returns nil.
// The debug argument is optional, and if set to true, will log a warning message if the transaction's 'to' address is nil.
func parseMethod(tx *types.Transaction, contractAbi abi.ABI, debug *bool) *DecodedMethod {
	// initialize an empty map to store method parameters
	params := make(map[string]interface{})

	// check if the transaction data is valid and contains at least 10 hex characters
	if len(string(tx.Data())) < 10 {
		return nil
	}

	// encode the transaction data as a hex string
	txData := hexutil.Encode(tx.Data())

	// extract the input data (excluding the first 2 characters) and the signature hash (first 8 characters)
	inputData := txData[10:]
	sigHash := txData[2:10]

	// convert the signature hash from hex string to bytes
	signatureBytes := common.FromHex(sigHash)

	// convert the input data from hex string to bytes
	inputBytes := common.Hex2Bytes(inputData)

	// find the method corresponding to the signature hash in the ABI
	method, err := contractAbi.MethodById(signatureBytes)

	// if there is an error or the method is not found, return nil
	if err != nil || method == nil {
		return nil
	}

	// unpack the method inputs into the params map
	err = method.Inputs.UnpackIntoMap(params, inputBytes)

	// if there is an error, log it and return nil
	if err != nil {
		log.Fatal(
			"error unpack method into map:", method.Name,
			">> hash:", tx.Hash().Hex(),
			">> input:", inputData,
			">> signature:", sigHash,
			">> error:", err,
		)
		return nil
	}

	// initialize the contract variable
	var contract string

	// if the transaction destination is not nil, set the contract to its address
	if tx.To() != nil {
		contract = tx.To().Hex()
	} else { // otherwise set it to a default address and log a warning if debug is enabled
		contract = "0x0000000000000000000000000000000000000000"
		if debug != nil && *debug {
			log.Fatal(`decoder: no tx.to in transaction:`, tx.Hash().String())
		}
	}

	// format the parameters and update the params map
	params = formatParameters(params, debug)

	// return the decoded method as a pointer to a DecodedMethod struct
	return &DecodedMethod{
		TransactionHash: tx.Hash().Hex(),
		Contract:        contract,
		SigHash:         sigHash,
		Signature:       method.Sig,
		Params:          params,
	}
}

// parseLog parses a Ethereum log entry and decodes its event parameters according to a given contract ABI.
// It returns a DecodedLog struct containing the decoded parameters, or nil if the log entry could not be decoded.
// vLog: the log entry to be decoded.
// contractAbi: the ABI of the contract where the log entry originated from.
// debug: if true, additional debug messages will be printed.
func parseLog(vLog *types.Log, contractAbi abi.ABI, debug *bool) *DecodedLog {
	// Check if the log entry has at least one topic (the event signature hash).
	if len(vLog.Topics) <= 0 {
		return nil
	}

	// Get the event corresponding to the signature hash.
	topic0 := vLog.Topics[0]
	params := make(map[string]interface{})
	event, err := contractAbi.EventByID(vLog.Topics[0])
	if err != nil {
		return nil
	}

	// Unpack the event parameters from the log data.
	err = contractAbi.UnpackIntoMap(params, event.Name, vLog.Data)
	if err != nil {
		// Some events may have different signatures than their ABI, or may contain invalid data.
		// If we cannot unpack the parameters, we check if the event is in a list of known skipped events,
		// or if the log data is empty. If so, we skip the event, otherwise we return nil.
		skip := []string{
			"Approval",
			"Transfer",
			"Deposit",
		}
		if !slices.Contains(skip, event.Name) {
			if hexutil.Encode(vLog.Data) != "0x" {
				return nil
			}
		} else {
			if debug != nil && *debug {
				log.Fatal(`unpack error`, event.Name, vLog.TxHash.String(), err)
			}
		}
	}

	// Decode indexed parameters by iterating through all inputs and looking for indexed values.
	if len(vLog.Topics) > 1 {
		idxIndexedTopics := 1
		for _, argument := range event.Inputs {
			if idxIndexedTopics >= len(vLog.Topics) {
				// Check if the number of indexed topics matches the expected number of inputs.
				var abi []byte
				contractAbi.UnmarshalJSON(abi)
				continue
			}
			if argument.Indexed {
				t := argument.Type.String()
				topicData := vLog.Topics[idxIndexedTopics]

				// Unpack the indexed parameter value and add it to the parameters map.
				value, err := contractAbi.Unpack(t, topicData.Bytes())
				if err != nil {
					if debug != nil && *debug {
						log.Fatal(fmt.Sprintf("failed to decode indexed parameter %s: %s\n", argument.Name, err))
					}
				} else {
					params[argument.Name] = value
				}
				idxIndexedTopics++
			}
		}
	}

	// Format the decoded parameters and return the DecodedLog struct.
	params = formatParameters(params, debug)
	return &DecodedLog{
		TransactionHash: vLog.TxHash.Hex(),
		LogIndex:        vLog.Index,
		Contract:        vLog.Address.Hex(),
		Topic:           topic0.Hex(),
		Signature:       event.Sig,
		Params:          params,
	}
}

// formatParameters will iterate through objects and will parse big.Int to string.
// it will also parse addresses and return them as checksum addresses.
func formatParameters(decoded map[string]interface{}, debug *bool) map[string]interface{} {
	for key, value := range decoded {
		switch value := value.(type) {
		// For *big.Int types, parse the value to string
		case *big.Int:
			decoded[key] = value.String()

		// For common.Address types, convert to a checksum address
		case common.Address:
			decoded[key] = value.Hex()

		// For [][]uint8 types, convert to a list of hex strings
		case [][]uint8:
			parsed := make([]string, 0, len(value))
			for _, arr := range value {
				parsed = append(parsed, "0x"+common.Bytes2Hex(arr))
			}
			decoded[key] = parsed

		// For []*big.Int types, convert to a list of strings
		case []*big.Int:
			parsed := make([]string, 0, len(value))
			for _, v := range value {
				parsed = append(parsed, v.String())
			}
			decoded[key] = parsed

		// For []common.Address types, convert to a list of checksum addresses
		case []common.Address:
			parsed := make([]string, 0, len(value))
			for _, address := range value {
				parsed = append(parsed, address.Hex())
			}
			decoded[key] = parsed

		// For []uint8 types, convert to a hex string
		case []uint8:
			decoded[key] = "0x" + common.Bytes2Hex(value)

		// For strings, booleans, and uint8 types, no parsing necessary
		case string, bool, uint8:
			// do nothing

		// For [32]uint8 types, convert to a checksum address
		case [32]uint8:
			ba := make([]byte, 0, 32)
			for _, b := range value {
				ba = append(ba, b)
			}
			decoded[key] = common.BytesToHash(ba).Hex()

		// For all other types, log a warning message if debug mode is enabled
		default:
			if debug != nil && *debug {
				log.Fatal(`key:`, key, `value:`, value, `type:`, reflect.TypeOf(value))
			}
		}

		// If debug mode is enabled, log the formatted value
		if debug != nil && *debug {
			log.Fatal(`formatted value:`, decoded[key])
		}
	}

	return decoded
}

func IsToken(bytecode string) bool {
	tr := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"[2:]
	return strings.Contains(bytecode, tr)
}

func IsERC721(bytecode string) bool {
	return IsToken(bytecode) && strings.Contains(bytecode, "6352211e")
}

func IsERC20(bytecode string) bool {
	return IsToken(bytecode) && strings.Contains(bytecode, "6352211e")
}

// helper function to detect token standard.
func detectTokenStandard(bytecode string) string {
	if IsToken(bytecode) && IsERC721(bytecode) {
		return "ERC721"
	}

	// Decimals + ttr
	if IsToken(bytecode) && IsERC20(bytecode) {
		return "ERC20"
	}

	return "UNKNOWN"
}

func DetectBytecodes(bytecode string, signatures []string) bool {
	found := 0
	for _, code := range signatures {
		if strings.Contains(bytecode, strings.Replace(code, "0x", "", -1)) {
			found++
		}
	}

	fmt.Println(len(signatures), found)

	return len(signatures) == found
}
