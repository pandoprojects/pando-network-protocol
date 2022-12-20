package tx

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/pandoprojects/pando/crypto"

	"github.com/pandoprojects/pando/crypto/bls"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/pandoprojects/pando/cmd/pandocli/cmd/utils"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/core"
	"github.com/pandoprojects/pando/ledger/types"
	"github.com/pandoprojects/pando/rpc"

	rpcc "github.com/ybbus/jsonrpc"
)

// depositStakeCmd represents the deposit stake command
// Example:
//		pandocli tx deposit --chain="privatenet" --source=2E833968E5bB786Ae419c4d13189fB081Cc43bab --holder=2E833968E5bB786Ae419c4d13189fB081Cc43bab --stake=6000000 --purpose=0 --seq=7
var depositStakeCmd = &cobra.Command{
	Use:     "deposit",
	Short:   "Deposit stake to a zyta or meta or rametronenterprise",
	Example: `pandocli tx deposit --chain="privatenet" --source=2E833968E5bB786Ae419c4d13189fB081Cc43bab --holder=2E833968E5bB786Ae419c4d13189fB081Cc43bab --stake=6000000 --purpose=0 --seq=7`,
	Run:     doDepositStakeCmd,
}

func doDepositStakeCmd(cmd *cobra.Command, args []string) {
	wallet, sourceAddress, err := walletUnlockWithPath(cmd, sourceFlag, pathFlag, passwordFlag)
	if err != nil {
		return
	}
	defer wallet.Lock(sourceAddress)

	fee, ok := types.ParseCoinAmount(feeFlag)
	if !ok {
		utils.Error("Failed to parse fee")
	}
	stake, ok := types.ParseCoinAmount(stakeInPandoFlag)
	if !ok {
		utils.Error("Failed to parse stake")
	}
	if stake.Cmp(core.Zero) < 0 {
		utils.Error("Invalid input: stake must be positive\n")
	}

	var pandoStake *big.Int
	var ptxStake *big.Int

	if purposeFlag == core.StakeForValidator || purposeFlag == core.StakeForGuardian {
		pandoStake = new(big.Int).SetUint64(0)
		ptxStake = stake
	} else { // purposeFlag == core.StakeForrametronenterprise
		pandoStake = new(big.Int).SetUint64(0)
		ptxStake = stake
	}

	source := types.TxInput{
		Address: sourceAddress,
		Coins: types.Coins{
			PandoWei: pandoStake,
			PTXWei: ptxStake,
		},
		Sequence: uint64(seqFlag),
	}

	depositStakeTx := &types.DepositStakeTxV2{
		Fee: types.Coins{
			PandoWei: pandoStake,
			PTXWei: fee,
		},
		Source:  source,
		Purpose: purposeFlag,
	}

	// Parse holder flag.
	var holderAddress common.Address
	if purposeFlag == core.StakeForValidator {
		if len(holderFlag) != 40 && len(holderFlag) != 42 {
			utils.Error("holder must be a valid address")
		}
		holderAddress = common.HexToAddress(holderFlag)
	} else if purposeFlag == core.StakeForGuardian {
		if strings.HasPrefix(holderFlag, "0x") {
			holderFlag = holderFlag[2:]
		}
		if len(holderFlag) != 458 {
			utils.Error("Holder must be a valid guardian summary")
		}
		guardianKeyBytes, err := hex.DecodeString(holderFlag)
		if err != nil {
			utils.Error("Failed to decode guardian address: %v\n", err)
		}
		holderAddress = common.BytesToAddress(guardianKeyBytes[:20])
		blsPubkey, err := bls.PublicKeyFromBytes(guardianKeyBytes[20:68])
		if err != nil {
			utils.Error("Failed to decode bls Pubkey: %v\n", err)
		}
		blsPop, err := bls.SignatureFromBytes(guardianKeyBytes[68:164])
		if err != nil {
			utils.Error("Failed to decode bls POP: %v\n", err)
		}
		holderSig, err := crypto.SignatureFromBytes(guardianKeyBytes[164:])
		if err != nil {
			utils.Error("Failed to decode signature: %v\n", err)
		}

		depositStakeTx.BlsPubkey = blsPubkey
		depositStakeTx.BlsPop = blsPop
		depositStakeTx.HolderSig = holderSig
	} else { // purposeFlag == core.StakeForrametronenterprise
		if strings.HasPrefix(holderFlag, "0x") {
			holderFlag = holderFlag[2:]
		}
		if len(holderFlag) != 458 {
			utils.Error("Holder must be a valid rametronenterprise summary")
		}
		rametronenterpriseKeyBytes, err := hex.DecodeString(holderFlag)
		if err != nil {
			utils.Error("Failed to decode rametronenterprise address: %v\n", err)
		}
		holderAddress = common.BytesToAddress(rametronenterpriseKeyBytes[:20])
		blsPubkey, err := bls.PublicKeyFromBytes(rametronenterpriseKeyBytes[20:68])
		if err != nil {
			utils.Error("Failed to decode bls Pubkey: %v\n", err)
		}
		blsPop, err := bls.SignatureFromBytes(rametronenterpriseKeyBytes[68:164])
		if err != nil {
			utils.Error("Failed to decode bls POP: %v\n", err)
		}
		holderSig, err := crypto.SignatureFromBytes(rametronenterpriseKeyBytes[164:])
		if err != nil {
			utils.Error("Failed to decode signature: %v\n", err)
		}

		depositStakeTx.BlsPubkey = blsPubkey
		depositStakeTx.BlsPop = blsPop
		depositStakeTx.HolderSig = holderSig
	}

	depositStakeTx.Holder = types.TxOutput{
		Address: holderAddress,
	}

	sig, err := wallet.Sign(sourceAddress, depositStakeTx.SignBytes(chainIDFlag))
	if err != nil {
		utils.Error("Failed to sign transaction: %v\n", err)
	}
	depositStakeTx.SetSignature(sourceAddress, sig)

	raw, err := types.TxToBytes(depositStakeTx)
	if err != nil {
		utils.Error("Failed to encode transaction: %v\n", err)
	}
	signedTx := hex.EncodeToString(raw)

	client := rpcc.NewRPCClient(viper.GetString(utils.CfgRemoteRPCEndpoint))

	var res *rpcc.RPCResponse
	if asyncFlag {
		res, err = client.Call("pando.BroadcastRawTransactionAsync", rpc.BroadcastRawTransactionArgs{TxBytes: signedTx})
	} else {
		res, err = client.Call("pando.BroadcastRawTransaction", rpc.BroadcastRawTransactionArgs{TxBytes: signedTx})
	}
	if err != nil {
		utils.Error("Failed to broadcast transaction: %v\n", err)
	}
	if res.Error != nil {
		utils.Error("Server returned error: %v\n", res.Error)
	}
	fmt.Printf("Successfully broadcasted transaction.\n")
}

func init() {
	depositStakeCmd.Flags().StringVar(&chainIDFlag, "chain", "", "Chain ID")
	depositStakeCmd.Flags().StringVar(&sourceFlag, "source", "", "Source of the stake")
	depositStakeCmd.Flags().StringVar(&holderFlag, "holder", "", "Holder of the stake")
	depositStakeCmd.Flags().StringVar(&pathFlag, "path", "", "Wallet derivation path")
	depositStakeCmd.Flags().StringVar(&feeFlag, "fee", fmt.Sprintf("%dwei", types.MinimumTransactionFeePTXWeiDec2022), "Fee")
	depositStakeCmd.Flags().Uint64Var(&seqFlag, "seq", 0, "Sequence number of the transaction")
	depositStakeCmd.Flags().StringVar(&stakeInPandoFlag, "stake", "5000000", "Pando amount to stake")
	depositStakeCmd.Flags().Uint8Var(&purposeFlag, "purpose", 0, "Purpose of staking")
	depositStakeCmd.Flags().StringVar(&walletFlag, "wallet", "soft", "Wallet type (soft|nano)")
	depositStakeCmd.Flags().BoolVar(&asyncFlag, "async", false, "block until tx has been included in the blockchain")
	depositStakeCmd.Flags().StringVar(&passwordFlag, "password", "", "password to unlock the wallet")

	depositStakeCmd.MarkFlagRequired("chain")
	depositStakeCmd.MarkFlagRequired("source")
	depositStakeCmd.MarkFlagRequired("holder")
	depositStakeCmd.MarkFlagRequired("seq")
	depositStakeCmd.MarkFlagRequired("stake")
}
