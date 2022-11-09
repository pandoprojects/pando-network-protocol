package rpc

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"github.com/spf13/viper"
	rpcc "github.com/ybbus/jsonrpc"

	"github.com/pandoprojects/pando/cmd/pandocli/cmd/utils"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/core"
	"github.com/pandoprojects/pando/ledger/types"
	trpc "github.com/pandoprojects/pando/rpc"
)

// ------------------------------- SendTx -----------------------------------

type SendArgs struct {
	ChainID  string `json:"chain_id"`
	From     string `json:"from"`
	To       string `json:"to"`
	PandoWei string `json:"pandowei"`
	PTXWei string `json:"ptxwei"`
	Fee      string `json:"fee"`
	Sequence string `json:"sequence"`
	Async    bool   `json:"async"`
}

type SendResult struct {
	TxHash string            `json:"hash"`
	Block  *core.BlockHeader `json:"block",rlp:"nil"`
}

func (t *PandoCliRPCService) Send(args *SendArgs, result *SendResult) (err error) {
	if len(args.From) == 0 || len(args.To) == 0 {
		return fmt.Errorf("The from and to address cannot be empty")
	}
	if args.From == args.To {
		return fmt.Errorf("The from and to address cannot be identical")
	}

	from := common.HexToAddress(args.From)
	to := common.HexToAddress(args.To)
	pandowei, ok := new(big.Int).SetString(args.PandoWei, 10)
	if !ok {
		return fmt.Errorf("Failed to parse pandowei: %v", args.PandoWei)
	}
	ptxwei, ok := new(big.Int).SetString(args.PTXWei, 10)
	if !ok {
		return fmt.Errorf("Failed to parse ptxwei: %v", args.PTXWei)
	}
	fee, ok := new(big.Int).SetString(args.Fee, 10)
	if !ok {
		return fmt.Errorf("Failed to parse fee: %v", args.Fee)
	}
	sequence, err := strconv.ParseUint(args.Sequence, 10, 64)
	if err != nil {
		return err
	}

	if !t.wallet.IsUnlocked(from) {
		return fmt.Errorf("The from address %v has not been unlocked yet", from.Hex())
	}

	inputs := []types.TxInput{{
		Address: from,
		Coins: types.Coins{
			PTXWei: new(big.Int).Add(ptxwei, fee),
			PandoWei: pandowei,
		},
		Sequence: sequence,
	}}
	outputs := []types.TxOutput{{
		Address: to,
		Coins: types.Coins{
			PTXWei: ptxwei,
			PandoWei: pandowei,
		},
	}}
	sendTx := &types.SendTx{
		Fee: types.Coins{
			PandoWei: new(big.Int).SetUint64(0),
			PTXWei: fee,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}

	signBytes := sendTx.SignBytes(args.ChainID)
	sig, err := t.wallet.Sign(from, signBytes)
	if err != nil {
		utils.Error("Failed to sign transaction: %v\n", err)
	}
	sendTx.SetSignature(from, sig)

	raw, err := types.TxToBytes(sendTx)
	if err != nil {
		utils.Error("Failed to encode transaction: %v\n", err)
	}
	signedTx := hex.EncodeToString(raw)

	client := rpcc.NewRPCClient(viper.GetString(utils.CfgRemoteRPCEndpoint))

	rpcMethod := "pando.BroadcastRawTransaction"
	if args.Async {
		rpcMethod = "pando.BroadcastRawTransactionAsync"
	}
	res, err := client.Call(rpcMethod, trpc.BroadcastRawTransactionArgs{TxBytes: signedTx})
	if err != nil {
		return err
	}
	if res.Error != nil {
		return fmt.Errorf("Server returned error: %v", res.Error)
	}
	trpcResult := &trpc.BroadcastRawTransactionResult{}
	err = res.GetObject(trpcResult)
	if err != nil {
		return fmt.Errorf("Failed to parse Pando node response: %v", err)
	}

	result.TxHash = trpcResult.TxHash
	result.Block = trpcResult.Block

	return nil
}
