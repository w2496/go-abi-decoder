package decoder

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestTokenStore(t *testing.T) {
	addr := common.HexToAddress(EtherAddress)
	TknStore.SetTknInfo(&ITknInfo{
		Address:   addr,
		IsERC20:   false,
		IsERC721:  false,
		IsERC1155: false,
		Name:      "empty-decoder",
		Symbol:    "EPTY",
		Decimals:  18,
		Meta:      "{}",
	})

	tkn, err := TknStore.GetTknInfo(addr)

	if err != nil {
		t.Fatal("error getting token info from store", err)
	}

	t.Log("Token loaded from TknStore", tkn.Name)

	dec := tkn.GetDecoder()
	if dec.Abi == nil {
		t.Fatal("error get decoder - no abi loaded")
	}

	if dec.ContractAddress == nil {
		t.Fatal("error getting token info from store - no contract address")
	}

	t.Log(dec.GetSigHashes())
}
