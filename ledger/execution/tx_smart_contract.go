package execution

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/pandoprojects/pando/blockchain"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/common/result"
	"github.com/pandoprojects/pando/core"
	"github.com/pandoprojects/pando/crypto"
	st "github.com/pandoprojects/pando/ledger/state"
	"github.com/pandoprojects/pando/ledger/types"
	"github.com/pandoprojects/pando/ledger/vm"
)

var _ TxExecutor = (*SmartContractTxExecutor)(nil)

// ------------------------------- SmartContractTx Transaction -----------------------------------

// SmartContractTxExecutor implements the TxExecutor interface
type SmartContractTxExecutor struct {
	state  *st.LedgerState
	chain  *blockchain.Chain
	ledger core.Ledger
}

// NewSmartContractTxExecutor creates a new instance of SmartContractTxExecutor
func NewSmartContractTxExecutor(chain *blockchain.Chain, state *st.LedgerState, ledger core.Ledger) *SmartContractTxExecutor {
	return &SmartContractTxExecutor{
		state:  state,
		chain:  chain,
		ledger: ledger,
	}
}

func (exec *SmartContractTxExecutor) sanityCheck(chainID string, view *st.StoreView, viewSel core.ViewSelector, transaction types.Tx) result.Result {
	blockHeight := getBlockHeight(exec.state)
	tx := transaction.(*types.SmartContractTx)

	// Validate from, basic
	res := tx.From.ValidateBasic()
	if res.IsError() {
		return res
	}

	// Check signatures
	signBytes := tx.SignBytes(chainID)
	nativeSignatureValid := tx.From.Signature.Verify(signBytes, tx.From.Address)
	if blockHeight >= common.HeightTxWrapperExtension {
		signBytesV2 := types.ChangeEthereumTxWrapper(signBytes, 2)
		nativeSignatureValid = nativeSignatureValid || tx.From.Signature.Verify(signBytesV2, tx.From.Address)
	}

	if !nativeSignatureValid {
		if blockHeight < common.HeightRPCCompatibility {
			return result.Error("Signature verification failed, SignBytes: %v",
				hex.EncodeToString(signBytes)).WithErrorCode(result.CodeInvalidSignature)
		}

		// interpret the signature as ETH tx signature
		if tx.From.Coins.PandoWei.Cmp(big.NewInt(0)) != 0 {
			return result.Error("Sending Pando with ETH transaction is not allowed") // extra check, since ETH transaction only signs the PTX part (i.e., value, gasPrice, gasLimit, etc)
		}

		ethSigningHash := tx.EthSigningHash(chainID, blockHeight)
		err := crypto.ValidateEthSignature(tx.From.Address, ethSigningHash, tx.From.Signature)
		if err != nil {
			return result.Error("ETH Signature verification failed, SignBytes: %v, error: %v",
				hex.EncodeToString(signBytes), err.Error()).WithErrorCode(result.CodeInvalidSignature)
		}
	}

	// Get input account
	fromAccount, success := getInput(view, tx.From)
	if success.IsError() {
		return result.Error("Failed to get the account (the address has no Pando nor PTX)")
	}

	// Validate input, advanced

	// Check sequence/coins
	seq, balance := fromAccount.Sequence, fromAccount.Balance
	if seq+1 != tx.From.Sequence {
		return result.Error("ValidateInputAdvanced: Got %v, expected %v. (acc.seq=%v)",
			tx.From.Sequence, seq+1, fromAccount.Sequence).WithErrorCode(result.CodeInvalidSequence)
	}

	// Check amount
	if !balance.IsGTE(tx.From.Coins) {
		return result.Error("Insufficient fund: balance is %v, tried to send %v",
			balance, tx.From.Coins).WithErrorCode(result.CodeInsufficientFund)
	}

	coins := tx.From.Coins.NoNil()
	if !coins.IsNonnegative() {
		return result.Error("Invalid value to transfer").
			WithErrorCode(result.CodeInvalidValueToTransfer)
	}

	if !sanityCheckForGasPrice(tx.GasPrice, blockHeight) {
		minimumGasPrice := types.GetMinimumGasPrice(blockHeight)
		return result.Error("Insufficient gas price. Gas price needs to be at least %v PTXWei", minimumGasPrice).
			WithErrorCode(result.CodeInvalidGasPrice)
	}

	maxGasLimit := types.GetMaxGasLimit(blockHeight)
	if new(big.Int).SetUint64(tx.GasLimit).Cmp(maxGasLimit) > 0 {
		return result.Error("Invalid gas limit. Gas limit needs to be at most %v", maxGasLimit).
			WithErrorCode(result.CodeInvalidGasLimit)
	}

	zero := big.NewInt(0)
	feeLimit := new(big.Int).Mul(tx.GasPrice, new(big.Int).SetUint64(tx.GasLimit))
	if feeLimit.BitLen() > 255 || feeLimit.Cmp(zero) < 0 {
		// There is no explicit upper limit for big.Int. Just be conservative
		// here to prevent potential overflow attack
		return result.Error("Fee limit too high").
			WithErrorCode(result.CodeFeeLimitTooHigh)
	}

	var minimalBalance types.Coins
	value := coins.PTXWei      // NoNil() already guarantees value is NOT nil
	pandoValue := coins.PandoWei // NoNil() already guarantees value is NOT nil
	if !vm.SupportPandoTransferInEVM(blockHeight) {
		minimalBalance = types.Coins{
			PandoWei: zero,
			PTXWei: feeLimit.Add(feeLimit, value),
		}
	} else {
		minimalBalance = types.Coins{
			PandoWei: pandoValue,
			PTXWei: feeLimit.Add(feeLimit, value),
		}
	}

	if !fromAccount.Balance.IsGTE(minimalBalance) {
		logger.Infof(fmt.Sprintf("Source did not have enough balance %v", tx.From.Address.Hex()))
		return result.Error("Source balance is %v, but required minimal balance is %v",
			fromAccount.Balance, minimalBalance).WithErrorCode(result.CodeInsufficientFund)
	}

	return result.OK
}

func (exec *SmartContractTxExecutor) process(chainID string, view *st.StoreView, viewSel core.ViewSelector, transaction types.Tx) (common.Hash, result.Result) {
	tx := transaction.(*types.SmartContractTx)

	view.ResetLogs()
	view.ResetBalanceChanges()

	// Note: for contract deployment, vm.Execute() might transfer coins from the fromAccount to the
	//       deployed smart contract. Thus, we should call vm.Execute() before calling getInput().
	//       Otherwise, the fromAccount returned by getInput() will have incorrect balance.
	evmRet, contractAddr, gasUsed, evmErr := vm.Execute(exec.state.ParentBlock(), tx, view)

	fromAddress := tx.From.Address
	fromAccount, success := getInput(view, tx.From)
	if success.IsError() {
		return common.Hash{}, result.Error("Failed to get the from account")
	}

	feeAmount := new(big.Int).Mul(tx.GasPrice, new(big.Int).SetUint64(gasUsed))
	fee := types.Coins{
		PandoWei: big.NewInt(int64(0)),
		PTXWei: feeAmount,
	}
	if !chargeFee(fromAccount, fee) {
		return common.Hash{}, result.Error("failed to charge transaction fee")
	}

	createContract := (tx.To.Address == common.Address{})
	if !createContract { // vm.create() increments the sequence of the from account
		fromAccount.Sequence++
	}
	view.SetAccount(fromAddress, fromAccount)

	txHash := types.TxID(chainID, tx)

	// TODO: Add tx receipt: status and events
	logs := view.PopLogs()
	balanceChanges := view.PopBalanceChanges()
	if evmErr != nil {
		// Do not record events if transaction is reverted
		logs = nil
		balanceChanges = nil
	}

	if viewSel == core.DeliveredView { // only record the receipt for the delivered views
		exec.chain.AddTxReceipt(exec.ledger.GetCurrentBlock(), tx, logs, balanceChanges, evmRet, contractAddr, gasUsed, evmErr)
	}

	return txHash, result.OK
}

func (exec *SmartContractTxExecutor) getTxInfo(transaction types.Tx) *core.TxInfo {
	tx := transaction.(*types.SmartContractTx)
	return &core.TxInfo{
		Address:           tx.From.Address,
		Sequence:          tx.From.Sequence,
		EffectiveGasPrice: exec.calculateEffectiveGasPrice(transaction),
	}
}

func (exec *SmartContractTxExecutor) calculateEffectiveGasPrice(transaction types.Tx) *big.Int {
	tx := transaction.(*types.SmartContractTx)
	return tx.GasPrice
}
