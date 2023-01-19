package common

// HeightEnableValidatorReward specifies the minimal block height to enable the validtor PTX reward
const HeightEnableValidatorReward uint64 = 1
// HeightEnablePando1 specifies the minimal block height to enable the Pando1.0 feature.
const HeightEnablePando1 uint64 = 1 //

// HeightLowermetatronodeStakeThresholdTo1000 specifies the minimal block height to lower the meta Stake Threshold to 1,000 Pando
const HeightLowerMetaStakeThresholdTo10000 uint64 = 4417900

// HeightEnableSmartContract specifies the minimal block height to eanble the Turing-complete smart contract support
const HeightEnableSmartContract uint64 = 1

// HeightSampleStakingReward specifies the block heigth to enable sampling of staking reward
const HeightSampleStakingReward uint64 = 1

// HeightJune2022FeeAdjustment specifies the block heigth to enable transaction fee burning adjustment
const HeightJune2022FeeAdjustment uint64 = 1

// HeightEnablePando2 specifies the minimal block height to enable the Pando2.0 feature.
const HeightEnablePando2 uint64 = 4417900

// HeightRPCCompatibility specifies the block height to enable Ethereum compatible RPC support
const HeightRPCCompatibility uint64 = 1

// HeightTxWrapperExtension specifies the block height to extend the Tx Wrapper
const HeightTxWrapperExtension uint64 = 1

// HeightSupportpandoprojectsInSmartContract specifies the block height to support Pando in smart contracts
const HeightSupportpandoprojectsInSmartContract uint64 = 4417900

// HeightZytaStakeChangedTo10000K specifies the block height to lower the validator stake to 200,000 Pando
const HeightZytaStakeChangedTo10000K uint64 = 4417900


// CheckpointInterval defines the interval between checkpoints.
const CheckpointInterval = int64(100)

// CheckpointInterval defines the interval between checkpoints.
const CheckpointIntervalForRametron = int64(1000)

// IsCheckPointHeight returns if a block height is a checkpoint.
func IsCheckPointHeight(height uint64) bool {
	return height%uint64(CheckpointInterval) == 1
}

// IsCheckPointHeight returns if a block height is a checkpoint.
func IsCheckPointHeightForRametron(height uint64) bool {
	return height%uint64(CheckpointIntervalForRametron) == 1
}


// LastCheckPointHeight returns the height of the last checkpoint
func LastCheckPointHeight(height uint64) uint64 {
	multiple := height / uint64(CheckpointInterval)
	lastCheckpointHeight := uint64(CheckpointInterval)*multiple + 1
	return lastCheckpointHeight
}


