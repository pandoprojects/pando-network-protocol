package types

import (
	"math/big"

	"github.com/pandoprojects/pando/common"
)

const (
	// DenomPandoWei is the basic unit of pando, 1 Pando = 10^18 PandoWei
	DenomPandoWei string = "PandoWei"

	// DenomPTXWei is the basic unit of pando, 1 Pando = 10^18 PandoWei
	DenomPTXWei string = "PTXWei"

	// Initial gas parameters

	// MinimumGasPrice is the minimum gas price for a smart contract transaction
	MinimumGasPrice uint64 = 1e8

	// MaximumTxGasLimit is the maximum gas limit for a smart contract transaction
	//MaximumTxGasLimit uint64 = 2e6
	MaximumTxGasLimit uint64 = 10e6

	// MinimumTransactionFeePTXWei specifies the minimum fee for a regular transaction
	MinimumTransactionFeePTXWei uint64 = 1e12

	// June 2021 gas burn adjustment

	// MinimumGasPrice is the minimum gas price for a smart contract transaction
	MinimumGasPriceDec2022 uint64 = 4e12
	
	// MaximumTxGasLimit is the maximum gas limit for a smart contract transaction
	MaximumTxGasLimitDec2022 uint64 = 20e6

	// MinimumTransactionFeePTXWei specifies the minimum fee for a regular transaction
	MinimumTransactionFeePTXWeiDec2022 uint64 = 1e15
	
	// MaxAccountsAffectedPerTx specifies the max number of accounts one transaction is allowed to modify to avoid spamming
	MaxAccountsAffectedPerTx = 512
)

const (
	// ValidatorPandoGenerationRateNumerator is used for calculating the generation rate of Pando for validators
	//ValidatorPandoGenerationRateNumerator int64 = 317
	ValidatorPandoGenerationRateNumerator int64 = 0 // ZERO inflation for Pando

	// ValidatorPandoGenerationRateDenominator is used for calculating the generation rate of Pando for validators
	// ValidatorPandoGenerationRateNumerator / ValidatorPandoGenerationRateDenominator is the amount of PandoWei
	// generated per existing PandoWei per new block
	ValidatorPandoGenerationRateDenominator int64 = 1e11

	// ValidatorPTXGenerationRateNumerator is used for calculating the generation rate of PTX for validators
	ValidatorPTXGenerationRateNumerator int64 = 0 // ZERO initial inflation for PTX

	// ValidatorPTXGenerationRateDenominator is used for calculating the generation rate of PTX for validators
	// ValidatorPTXGenerationRateNumerator / ValidatorPTXGenerationRateDenominator is the amount of PTXWei
	// generated per existing PandoWei per new block
	ValidatorPTXGenerationRateDenominator int64 = 1e9

	// RegularPTXGenerationRateNumerator is used for calculating the generation rate of PTX for other types of accounts
	//RegularPTXGenerationRateNumerator int64 = 1900
	RegularPTXGenerationRateNumerator int64 = 0 // ZERO initial inflation for PTX

	// RegularPTXGenerationRateDenominator is used for calculating the generation rate of PTX for other types of accounts
	// RegularPTXGenerationRateNumerator / RegularPTXGenerationRateDenominator is the amount of PTXWei
	// generated per existing PandoWei per new block
	RegularPTXGenerationRateDenominator int64 = 1e10
)

const (

	// ServiceRewardVerificationBlockDelay gives the block delay for service certificate verification
	ServiceRewardVerificationBlockDelay uint64 = 2

	// ServiceRewardFulfillmentBlockDelay gives the block delay for service reward fulfillment
	ServiceRewardFulfillmentBlockDelay uint64 = 4
)

const (

	// MaximumTargetAddressesForStakeBinding gives the maximum number of target addresses that can be associated with a bound stake
	MaximumTargetAddressesForStakeBinding uint = 1024

	// MaximumFundReserveDuration indicates the maximum duration (in terms of number of blocks) of reserving fund
	MaximumFundReserveDuration uint64 = 12 * 3600

	// MinimumFundReserveDuration indicates the minimum duration (in terms of number of blocks) of reserving fund
	MinimumFundReserveDuration uint64 = 300

	// ReservedFundFreezePeriodDuration indicates the freeze duration (in terms of number of blocks) of the reserved fund
	ReservedFundFreezePeriodDuration uint64 = 5
)

func GetMinimumGasPrice(blockHeight uint64) *big.Int {
	if blockHeight < common.HeightJune2022FeeAdjustment {
		return new(big.Int).SetUint64(MinimumGasPrice)
	}

	return new(big.Int).SetUint64(MinimumGasPriceDec2022)
}

func GetMaxGasLimit(blockHeight uint64) *big.Int {
	if blockHeight < common.HeightJune2022FeeAdjustment {
		return new(big.Int).SetUint64(MaximumTxGasLimit)
	}

	return new(big.Int).SetUint64(MaximumTxGasLimitDec2022)
}

func GetMinimumTransactionFeePTXWei(blockHeight uint64) *big.Int {
	if blockHeight < common.HeightJune2022FeeAdjustment {
		return new(big.Int).SetUint64(MinimumTransactionFeePTXWei)
	}

	return new(big.Int).SetUint64(MinimumTransactionFeePTXWeiDec2022)
}

// Special handling for many-to-many SendTx
func GetSendTxMinimumTransactionFeePTXWei(numAccountsAffected uint64, blockHeight uint64) *big.Int {
	if blockHeight < common.HeightJune2022FeeAdjustment {
		return new(big.Int).SetUint64(MinimumTransactionFeePTXWei) // backward compatiblity
	}

	if numAccountsAffected < 2 {
		numAccountsAffected = 2
	}

	// minSendTxFee = numAccountsAffected * MinimumTransactionFeePTXWeiDec2022 / 2
	minSendTxFee := big.NewInt(1).Mul(new(big.Int).SetUint64(numAccountsAffected), new(big.Int).SetUint64(MinimumTransactionFeePTXWeiDec2022))
	minSendTxFee = big.NewInt(1).Div(minSendTxFee, new(big.Int).SetUint64(2))

	return minSendTxFee
}
