package decoder

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ctxType struct {
	connection  *string
	initialized bool
	isLegacy    *bool
	chainId     *big.Int
	signer      types.Signer
	eth         *ethclient.Client
}

var Ctx = ctxType{
	initialized: false,
}

// NewCtx is an initializer function for TxnSign.
func NewCtx(chainId *big.Int) ctxType {
	if Ctx.eth == nil {
		return ctxType{
			initialized: false,
			isLegacy:    nil,
			chainId:     nil,
		}
	}

	if chainId != nil && Ctx.chainId == chainId && Ctx.initialized {
		return ctxType{
			initialized: Ctx.initialized,
			connection:  Ctx.connection,
			isLegacy:    Ctx.isLegacy,
			chainId:     Ctx.chainId,
			signer:      Ctx.signer,
			eth:         Ctx.eth,
		}
	}

	var signer types.Signer
	ctx := context.Background()

	if chainId == nil && Ctx.chainId == nil {
		id, err := Ctx.eth.ChainID(ctx)
		if err == nil && id != nil {
			Ctx.chainId = id
			chainId = id
		}
	} else if chainId == nil && Ctx.chainId != nil {
		chainId = Ctx.chainId
	}

	if Ctx.chainId == nil || Ctx.chainId != chainId {
		Ctx.chainId = chainId
	}

	if Ctx.isLegacy == nil {
		is, err := IsEIP1559(Ctx.eth, ctx)
		if err == nil && is != nil {
			Ctx.isLegacy = is
		}
	}

	if Ctx.isLegacy != nil && *Ctx.isLegacy {
		signer = types.NewEIP155Signer(chainId)
	} else {
		signer = types.NewLondonSigner(chainId)
	}

	Ctx.initialized = true

	return ctxType{
		isLegacy: Ctx.isLegacy, eth: Ctx.eth,
		chainId: chainId, signer: signer,
	}
}

func SetClient(client *ethclient.Client) *ethclient.Client {
	Ctx.eth = client
	Ctx = NewCtx(nil)
	return Ctx.eth
}

func GetClient() *ethclient.Client {
	return Ctx.eth
}

func Connect(nodeUrl string) *ethclient.Client {
	Ctx.connection = &nodeUrl
	client, err := ethclient.Dial(nodeUrl)

	if err != nil {
		panic(err)
	}

	return SetClient(client)
}

func (s *ctxType) ReloadCtx() {
	Ctx = NewCtx(Ctx.chainId)
}

func (s *ctxType) GetTxFrom(tx *types.Transaction) *string {
	if from, err := types.Sender(s.signer, tx); err == nil {
		sender := from.Hex()
		return &sender
	}

	return nil
}

func (*ctxType) GetMinerAndNonce(block *types.Block) (miner string, nonce string) {
	return GetMinerAndNonce(block)
}
