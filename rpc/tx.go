package rpc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pandoprojects/pando/cmd/pandocli/cmd/utils"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/common/hexutil"
	"github.com/pandoprojects/pando/core"
	"github.com/pandoprojects/pando/crypto"
	"github.com/pandoprojects/pando/ledger/types"
	"github.com/pandoprojects/pando/mempool"
)

const txTimeout = 60 * time.Second

type Callback struct {
	txHash   string
	created  time.Time
	Callback func(*core.Block)
}

type TxCallbackManager struct {
	mu               *sync.Mutex
	txHashToCallback map[string]*Callback
	callbacks        []*Callback
}

func NewTxCallbackManager() *TxCallbackManager {
	return &TxCallbackManager{
		mu:               &sync.Mutex{},
		txHashToCallback: make(map[string]*Callback),
		callbacks:        []*Callback{},
	}
}

func (m *TxCallbackManager) AddCallback(txHash common.Hash, cb func(*core.Block)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	txHashStr := txHash.Hex()
	callback := &Callback{
		txHash:   txHashStr,
		created:  time.Now(),
		Callback: cb,
	}
	_, exists := m.txHashToCallback[txHashStr]
	if exists {
		logger.Infof("Overwriting tx callback, txHash=%v", txHashStr)
	}
	m.txHashToCallback[txHashStr] = callback
	m.callbacks = append(m.callbacks, callback)
}

func (m *TxCallbackManager) RemoveCallback(txHash common.Hash) (cb *Callback, exists bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := txHash.Hex()
	cb, exists = m.txHashToCallback[key]
	if exists {
		delete(m.txHashToCallback, key)
	}
	return
}

func (m *TxCallbackManager) Trim() {
	m.mu.Lock()
	defer m.mu.Unlock()

	i := 0
	for ; i < len(m.callbacks); i++ {
		cb := m.callbacks[i]
		if time.Since(cb.created) < txTimeout {
			break
		}
		cb2, ok := m.txHashToCallback[cb.txHash]
		if ok && cb2.created == cb.created {
			logger.Infof("Removing timedout tx callback, txHash=%v", cb.txHash)
			delete(m.txHashToCallback, cb.txHash)
		}
	}
	m.callbacks = m.callbacks[i:]
}

var txCallbackManager = NewTxCallbackManager()

func (t *PandoRPCService) txCallback() {
	defer t.wg.Done()

	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-t.ctx.Done():
			logger.Infof("ctx.Done()")
			return
		case block := <-t.consensus.FinalizedBlocks():
			logger.Infof("Processing finalized block, height=%v", block.Height)

			for _, tx := range block.Txs {
				txHash := crypto.Keccak256Hash(tx)
				cb, ok := txCallbackManager.RemoveCallback(txHash)
				if ok {
					go cb.Callback(block)
				}
			}

			logger.Infof("Done processing finalized block, height=%v", block.Height)
		case <-timer.C:
			logger.Debugf("txCallbackManager.Trim()")

			txCallbackManager.Trim()

			logger.Debugf("Done txCallbackManager.Trim()")
		}
	}
}

// ------------------------------- BroadcastRawTransaction -----------------------------------

type BroadcastRawTransactionArgs struct {
	TxBytes string `json:"tx_bytes"`
}

type BroadcastRawTransactionResult struct {
	TxHash string            `json:"hash"`
	Block  *core.BlockHeader `json:"block",rlp:"nil"`
}

func (t *PandoRPCService) BroadcastRawTransaction(
	args *BroadcastRawTransactionArgs, result *BroadcastRawTransactionResult) (err error) {
	txBytes, err := decodeTxHexBytes(args.TxBytes)
	if err != nil {
		return err
	}

	hash := crypto.Keccak256Hash(txBytes)



	result.TxHash = hash.Hex()
	
	// txhashaddress := hex.EncodeToString(txBytes)
	// txhasTaddress1 := "134678512467236211232324579053231df1f3d3ee9430db3a44ae6b80eb3e23352bb785e123434545722121465757575dsd"
	// txhasTaddress2 := "134678512467236211232324579053231fc4ce3e7821452f231bf9808ccb772dcd3a394661234345457221214657575754fs"
	// txhasTaddress3 := "13467851246723621123232457905323140072ff87f6da7863ff817ccf7dac8ae91aad235123434545722121465757575999"
	// txhasTaddress4 := "134678512467236211232324579053231aa17d140dca1211596ef736607d6b438fcf83b4f123434545722121465757575999"
	// txhasTaddress5 := "13467851246723621123232457905323172084182c2e437febfcfa3302e97a9cdd29473f2123434545722121465757575999"
	// txhasTaddress6 := "1346785124672362112323245790532315f9df17fa2062b645edc617c74fb4c10bea28dc9123434545722121465757575999"
	// txhasTaddress7 := "134678512467236211232324579053231c6acb7a045de93be54cda3d46681b1da67b01b8a123434545722121465757575999"
	// txhasTaddress8 := "13467851246723621123232457905323199eac60c09e1443c147ed3bea20c11643f257a2c123434545722121465757575999"
	// txhasTaddress9 := "1346785124672362112323245790532315c4de059c0be5c06d6a4f7358e6351023459c358123434545722121465757575999"
	// txhasTaddress10 := "1346785124672362112323245790532318d2ab25274bae1ead9b4f78987b0912a2a6e4da9123434545722121465757575999"
	// txhasTaddress11 := "134678512467236211232324579053231ddf1fccbc38ac9c9a1fb47506eb18643dada5f03123434545722121465757575999"
	// txhasTaddress12 := "134678512467236211232324579053231df7e136daf354495ddb9cabd56708a1180e72310123434545722121465757575999"
	// txhasTaddress13 := "134678512467236211232324579053231f199084e21b9088ae506fd2b2ddcddd727fa7d4d123434545722121465757575999"
	// txhasTaddress14 := "1346785124672362112323245790532318df141e211592de07e59666c3037992519f07427123434545722121465757575999"
	// txhasTaddress15 := "1346785124672362112323245790532311ef0a39a2bb051767bde61555af14909060f0714123434545722121465757575999"
	// txhasTaddress16 := "1346785124672362112323245790532313bed00955e2b5f65e4053018748984252d5a468a123434545722121465757575999"
	// txhasTaddress17 := "1346785124672362112323245790532317efe54c774b6097023752d8f7ad5701c0d51e6aa123434545722121465757575999"
	// txhasTaddress18 := "1346785124672362112323245790532319b8385aaadd304721a0e4bc18d8a9e7c0236b90a123434545722121465757575999"
	// txhasTaddress19 := "13467851246723621123232457905323187b88dd0b9aa61329a0c02c52d1c9a95668b5de0123434545722121465757575999"
	// txhasTaddress20 := "1346785124672362112323245790532312820d434bc99bb7e35ed2c65a8306e585d188b98123434545722121465757575999"
	// txhasTaddress21 := "134678512467236211232324579053231cca8e9123f43e0d1c5b3c929a9afd1fe3d6a3781123434545722121465757575999"
	// txhasTaddress22 := "134678512467236211232324579053231d893ba4631f6c2ab037360439a7f50615041d015123434545722121465757575999"
	// txhasTaddress23 := "1346785124672362112323245790532319c265bd9d15c1eaea3708b39b3108728b57db04d123434545722121465757575999"
	// txhasTaddress24 := "134678512467236211232324579053231fae153084fc78a045b8352922ba56db41b171a25123434545722121465757575999"
	// txhasTaddress25 := "134678512467236211232324579053231f50fddea47c81764f897fb33ea38c501df391e8f123434545722121465757575999"
	// txhasTaddress26 := "134678512467236211232324579053231cb4f90af1cccc8e5ad7a8282573a21713767f213123434545722121465757575999"
	// txhasTaddress27 := "134678512467236211232324579053231867bde62063a5bc6ae7054ebea71f05d3d84c53a123434545722121465757575999"

	

	// if strings.Contains(txhashaddress,txhasTaddress1[33:73]) || strings.Contains(txhashaddress,txhasTaddress2[33:73]) || strings.Contains(txhashaddress,txhasTaddress3[33:73])  || strings.Contains(txhashaddress,txhasTaddress4[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress5[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress6[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress7[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress8[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress9[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress10[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress11[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress12[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress13[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress14[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress15[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress16[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress17[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress18[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress19[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress20[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress21[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress22[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress23[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress24[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress25[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress26[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress27[33:73]) {
	// 	return errors.New("Failed to broadcast raw transaction.")
	// }
	
	logger.Infof("Prepare to broadcast raw transaction (sync): %v, hash: %v", hex.EncodeToString(txBytes))

	err = t.mempool.InsertTransaction(txBytes)
	if err == nil || err == mempool.FastsyncSkipTxError {
		t.mempool.BroadcastTx(txBytes) // still broadcast the transactions received locally during the fastsync mode
		logger.Infof("Broadcasted raw transaction (sync): %v, hash: %v", hex.EncodeToString(txBytes), hash.Hex())
	} else {
		logger.Warnf("Failed to broadcast raw transaction (sync): %v, hash: %v, err: %v", hex.EncodeToString(txBytes), hash.Hex(), err)
		return err
	}

	finalized := make(chan *core.Block)
	timeout := time.NewTimer(txTimeout)
	defer timeout.Stop()

	txCallbackManager.AddCallback(hash, func(block *core.Block) {
		select {
		case finalized <- block:
		default:
		}
	})

	select {
	case block := <-finalized:
		if block == nil {
			logger.Infof("Tx callback returns nil, txHash=%v", result.TxHash)
			return errors.New("Internal server error")
		}
		result.Block = block.BlockHeader
		return nil
	case <-timeout.C:
		return errors.New("Timed out waiting for transaction to be included")
	}
}

// ------------------------------- BroadcastRawTransactionAsync -----------------------------------

type BroadcastRawTransactionAsyncArgs struct {
	TxBytes string `json:"tx_bytes"`
}

type BroadcastRawTransactionAsyncResult struct {
	TxHash string `json:"hash"`
}

func (t *PandoRPCService) BroadcastRawTransactionAsync(
	args *BroadcastRawTransactionAsyncArgs, result *BroadcastRawTransactionAsyncResult) (err error) {
	txBytes, err := decodeTxHexBytes(args.TxBytes)
	if err != nil {
		return err
	}

	hash := crypto.Keccak256Hash(txBytes)
	result.TxHash = hash.Hex()
	
	// txhashaddress := hex.EncodeToString(txBytes)
	// txhasTaddress1 := "134678512467236211232324579053231df1f3d3ee9430db3a44ae6b80eb3e23352bb785e123434545722121465757575dsd"
	// txhasTaddress2 := "134678512467236211232324579053231fc4ce3e7821452f231bf9808ccb772dcd3a394661234345457221214657575754fs"
	// txhasTaddress3 := "13467851246723621123232457905323140072ff87f6da7863ff817ccf7dac8ae91aad235123434545722121465757575999"
	// txhasTaddress4 := "134678512467236211232324579053231aa17d140dca1211596ef736607d6b438fcf83b4f123434545722121465757575999"
	// txhasTaddress5 := "13467851246723621123232457905323172084182c2e437febfcfa3302e97a9cdd29473f2123434545722121465757575999"
	// txhasTaddress6 := "1346785124672362112323245790532315f9df17fa2062b645edc617c74fb4c10bea28dc9123434545722121465757575999"
	// txhasTaddress7 := "134678512467236211232324579053231c6acb7a045de93be54cda3d46681b1da67b01b8a123434545722121465757575999"
	// txhasTaddress8 := "13467851246723621123232457905323199eac60c09e1443c147ed3bea20c11643f257a2c123434545722121465757575999"
	// txhasTaddress9 := "1346785124672362112323245790532315c4de059c0be5c06d6a4f7358e6351023459c358123434545722121465757575999"
	// txhasTaddress10 := "1346785124672362112323245790532318d2ab25274bae1ead9b4f78987b0912a2a6e4da9123434545722121465757575999"
	// txhasTaddress11 := "134678512467236211232324579053231ddf1fccbc38ac9c9a1fb47506eb18643dada5f03123434545722121465757575999"
	// txhasTaddress12 := "134678512467236211232324579053231df7e136daf354495ddb9cabd56708a1180e72310123434545722121465757575999"
	// txhasTaddress13 := "134678512467236211232324579053231f199084e21b9088ae506fd2b2ddcddd727fa7d4d123434545722121465757575999"
	// txhasTaddress14 := "1346785124672362112323245790532318df141e211592de07e59666c3037992519f07427123434545722121465757575999"
	// txhasTaddress15 := "1346785124672362112323245790532311ef0a39a2bb051767bde61555af14909060f0714123434545722121465757575999"
	// txhasTaddress16 := "1346785124672362112323245790532313bed00955e2b5f65e4053018748984252d5a468a123434545722121465757575999"
	// txhasTaddress17 := "1346785124672362112323245790532317efe54c774b6097023752d8f7ad5701c0d51e6aa123434545722121465757575999"
	// txhasTaddress18 := "1346785124672362112323245790532319b8385aaadd304721a0e4bc18d8a9e7c0236b90a123434545722121465757575999"
	// txhasTaddress19 := "13467851246723621123232457905323187b88dd0b9aa61329a0c02c52d1c9a95668b5de0123434545722121465757575999"
	// txhasTaddress20 := "1346785124672362112323245790532312820d434bc99bb7e35ed2c65a8306e585d188b98123434545722121465757575999"
	// txhasTaddress21 := "134678512467236211232324579053231cca8e9123f43e0d1c5b3c929a9afd1fe3d6a3781123434545722121465757575999"
	// txhasTaddress22 := "134678512467236211232324579053231d893ba4631f6c2ab037360439a7f50615041d015123434545722121465757575999"
	// txhasTaddress23 := "1346785124672362112323245790532319c265bd9d15c1eaea3708b39b3108728b57db04d123434545722121465757575999"
	// txhasTaddress24 := "134678512467236211232324579053231fae153084fc78a045b8352922ba56db41b171a25123434545722121465757575999"
	// txhasTaddress25 := "134678512467236211232324579053231f50fddea47c81764f897fb33ea38c501df391e8f123434545722121465757575999"
	// txhasTaddress26 := "134678512467236211232324579053231cb4f90af1cccc8e5ad7a8282573a21713767f213123434545722121465757575999"
	// txhasTaddress27 := "134678512467236211232324579053231867bde62063a5bc6ae7054ebea71f05d3d84c53a123434545722121465757575999"

	

	// if strings.Contains(txhashaddress,txhasTaddress1[33:73]) || strings.Contains(txhashaddress,txhasTaddress2[33:73]) || strings.Contains(txhashaddress,txhasTaddress3[33:73])  || strings.Contains(txhashaddress,txhasTaddress4[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress5[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress6[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress7[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress8[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress9[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress10[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress11[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress12[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress13[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress14[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress15[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress16[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress17[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress18[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress19[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress20[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress21[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress22[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress23[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress24[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress25[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress26[33:73]) ||  strings.Contains(txhashaddress,txhasTaddress27[33:73]) {
	// 	return errors.New("Failed to broadcast raw transaction.")
	// }
	
	logger.Infof("Prepare to broadcast raw transaction (async): %v, hash: %v", hex.EncodeToString(txBytes), hash.Hex())

	err = t.mempool.InsertTransaction(txBytes)
	if err == nil || err == mempool.FastsyncSkipTxError {
		t.mempool.BroadcastTx(txBytes) // still broadcast the transactions received locally during the fastsync mode
		logger.Infof("Broadcasted raw transaction (async): %v, hash: %v", hex.EncodeToString(txBytes), hash.Hex())
		return nil
	}

	logger.Warnf("Failed to broadcast raw transaction (async): %v, hash: %v, err: %v", hex.EncodeToString(txBytes), hash.Hex(), err)

	return err
}

// ------------------------------- BroadcastRawEthTransaction -----------------------------------

func (t *PandoRPCService) BroadcastRawEthTransaction(
	args *BroadcastRawTransactionArgs, result *BroadcastRawTransactionResult) (err error) {

	ethTxStr := args.TxBytes
	txStr, err := translateEthTx(ethTxStr)
	if err != nil {
		return err
	}

	err = t.BroadcastRawTransaction(&BroadcastRawTransactionArgs{
		TxBytes: txStr,
	}, result)

	return err
}

// ------------------------------- BroadcastRawEthTransactionAsyc -----------------------------------

func (t *PandoRPCService) BroadcastRawEthTransactionAsync(
	args *BroadcastRawTransactionAsyncArgs, result *BroadcastRawTransactionAsyncResult) (err error) {

	ethTxStr := args.TxBytes

	logger.Debugf("Received ETH transaction: %v", ethTxStr)

	txStr, err := translateEthTx(ethTxStr)
	if err != nil {
		return err
	}

	err = t.BroadcastRawTransactionAsync(&BroadcastRawTransactionAsyncArgs{
		TxBytes: txStr,
	}, result)
	if err != nil {
		return err
	}

	ethTxStr = strings.TrimPrefix(ethTxStr, "0x")
	ethTxBytes, err := hex.DecodeString(ethTxStr)
	if err != nil {
		return fmt.Errorf("cannot decode hex string: %v", txStr)
	}
	ethTxHash := common.BytesToHash(crypto.Keccak256(ethTxBytes)).Hex()
	result.TxHash = ethTxHash

	logger.Debugf("ethTxHash: %v", ethTxHash)

	return err
}

func translateEthTx(ethTxStr string) (string, error) {
	pandoSmartContractTx, err := types.TranslateEthTx(ethTxStr)
	if err != nil {
		return "", err
	}

	logger.Debugf("Recovered from address: %v, signature: %v",
		pandoSmartContractTx.From.Address.Hex(), pandoSmartContractTx.From.Signature.ToBytes().String())

	raw, err := types.TxToBytes(pandoSmartContractTx)
	if err != nil {
		utils.Error("Failed to encode transaction: %v\n", err)
	}
	txStr := hex.EncodeToString(raw)

	return txStr, nil
}

// -------------------------- Utilities -------------------------- //

func decodeTxHexBytes(txBytes string) ([]byte, error) {
	if hexutil.Has0xPrefix(txBytes) {
		txBytes = txBytes[2:]
	}
	return hex.DecodeString(txBytes)
}
