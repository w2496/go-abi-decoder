package decoder

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/exp/slices"
)

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
		contract = EtherAddress
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
		SigHash:         "0x" + sigHash,
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
	params := Params{}
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
				fmt.Println("ERROR UNPACK LOG DATA", err, event.Name)
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

					td := topicData.String()
					if td[0:26] == "0x000000000000000000000000" {
						params[argument.Name] = common.HexToAddress(topicData.String()).Hex()
						if debug != nil && *debug {
							fmt.Printf(`key: %v - value: %v\n`, argument.Name, params[argument.Name])
						}
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
		BlockNumber:     vLog.BlockNumber,
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
func formatParameters(decoded map[string]interface{}, debug *bool) Params {
	for key, value := range decoded {
		switch value := value.(type) {
		// For *big.Int types, parse the value to string
		case *big.Int:
			decoded[key] = value.String()

		// For common.Address types, convert to a checksum address
		case *common.Address:
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
		// for strings we check for address and checksum it
		case string:
			if value != EtherAddress && common.IsHexAddress(value) {
				decoded[key] = common.HexToAddress(value).Hex()
			}
		// For booleans, and uint8 types, no parsing necessary
		case bool, uint8:
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

func getBytecode(address common.Address) *string {
	if Ctx.eth == nil {
		return nil
	}

	code, err := Ctx.eth.CodeAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal("error getting bytecode:", address, err)
		zeroHex := "0x"
		return &zeroHex
	}

	res := strings.Join([]string{"0x", common.Bytes2Hex(code)}, "")
	return &res
}

func clientRequired() error {
	if Ctx.eth == nil && Ctx.connection == nil {
		return fmt.Errorf("no client connected or connection string attached to decoder.Ctx.eth")
	}

	if Ctx.eth == nil && Ctx.connection != nil {
		Connect(*Ctx.connection)
	}

	return nil
}

func getSymbol(ctx context.Context, contract common.Address) *string {
	if err := clientRequired(); err != nil {
		return nil
	}

	msg := ethereum.CallMsg{
		To: &contract, Data: common.Hex2Bytes("95d89b41"),
	}
	symbol, err := Ctx.eth.CallContract(ctx, msg, nil)

	if err != nil {
		return nil
	}

	result := ToAscii(symbol)

	return &result
}

func getName(ctx context.Context, contract common.Address) *string {
	if err := clientRequired(); err != nil {
		return nil
	}

	msg := ethereum.CallMsg{
		To: &contract, Data: common.Hex2Bytes("06fdde03"),
	}

	name, err := Ctx.eth.CallContract(ctx, msg, nil)
	if err != nil {
		return nil
	}

	out0 := ToAscii(name)

	return &out0
}

func getDecimals(ctx context.Context, contract common.Address) *uint8 {
	if err := clientRequired(); err != nil {
		return nil
	}

	msg := ethereum.CallMsg{
		To: &contract, Data: common.Hex2Bytes("313ce567"),
	}
	decimals, err := Ctx.eth.CallContract(ctx, msg, nil)

	if err != nil {
		return nil
	}

	result := uint8(common.BytesToHash(decimals).Big().Uint64())
	return &result
}

func getERC20Balance(ctx context.Context, address common.Address, contractAddress common.Address) (uint64, error) {
	if err := clientRequired(); err != nil {
		return 0, err
	}

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
	result, err := Ctx.eth.CallContract(ctx, msg, nil)
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

func queryTokenInfo(ctx context.Context, address common.Address, bytecodes ...string) ITknInfo {
	var code *string
	if len(bytecodes) > 0 {
		var byteSlice []string
		for _, bc := range bytecodes {
			byteSlice = append(byteSlice, strings.TrimPrefix(bc, "0x"))
		}

		joined := "0x" + strings.TrimPrefix(strings.Join(byteSlice, ""), "0x")
		code = &joined

	} else {
		code = getBytecode(address)
	}

	symbol := getSymbol(ctx, address)
	name := getName(ctx, address)
	decimals := getDecimals(ctx, address)

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
