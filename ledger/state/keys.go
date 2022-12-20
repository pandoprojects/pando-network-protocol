package state

import (
	"strconv"

	"github.com/pandoprojects/pando/common"
)

//
// ------------------------- Ledger State Keys -------------------------
//

// ChainIDKey returns the key for chainID
func ChainIDKey() common.Bytes {
	return common.Bytes("chainid")
}

// AccountKey constructs the state key for the given address
func AccountKey(addr common.Address) common.Bytes {
	return append(common.Bytes("ls/a/"), addr[:]...)
}

// SplitRuleKeyPrefix returns the prefix for the split rule key
func SplitRuleKeyPrefix() common.Bytes {
	return common.Bytes("ls/ssc/split/") // special smart contract / split rule
}

// SplitRuleKey constructs the state key for the given resourceID
func SplitRuleKey(resourceID string) common.Bytes {
	resourceIDBytes := common.Bytes(resourceID)
	return append(SplitRuleKeyPrefix(), resourceIDBytes[:]...)
}

// CodeKey constructs the state key for the given code hash
func CodeKey(codeHash common.Bytes) common.Bytes {
	return append(common.Bytes("ls/ch/"), codeHash...)
}

// ValidatorCandidatePoolKey returns the state key for the validator stake holder set
func ValidatorCandidatePoolKey() common.Bytes {
	return common.Bytes("ls/vcp")
}

// GuardianCandidatePoolKey returns the state key for the guadian stake holder set
func GuardianCandidatePoolKey() common.Bytes {
	return common.Bytes("ls/gcp")
}

// // RametronenterprisePoolKey returns the state key for the rametronenterprise PTX stake holder set
// func RametronenterprisePoolKey() common.Bytes {
// 	return common.Bytes("ls/rametronenterprisep")
// }

// RametronenterpriseKeyPrefix returns the prefix of the rametronenterprise key
func RametronenterpriseKeyPrefix() common.Bytes {
	return common.Bytes("ls/rametronenterprise/")
}

// RametronenterpriseKey returns the rametronenterprise key of a given address
func RametronenterpriseKey(addr common.Address) common.Bytes {
	prefix := RametronenterpriseKeyPrefix()
	return append(prefix, addr[:]...)
}

// StakeTransactionHeightListKey returns the state key the heights of blocks
// that contain stake related transactions (i.e. StakeDeposit, StakeWithdraw, etc)
func StakeTransactionHeightListKey() common.Bytes {
	return common.Bytes("ls/sthl")
}

// StatePruningProgressKey returns the key for the state pruning progress
func StatePruningProgressKey() common.Bytes {
	return common.Bytes("ls/spp")
}

// StakeRewardDistributionRuleSetKeyPrefix returns the prefix of the stake reward distribution rule
func StakeRewardDistributionRuleSetKeyPrefix() common.Bytes {
	return common.Bytes("ls/srdrs/")
}

// StakeRewardDistributionRuleSetKey returns the prefix of the stake reward distribution rule
func StakeRewardDistributionRuleSetKey(addr common.Address) common.Bytes {
	prefix := StakeRewardDistributionRuleSetKeyPrefix()
	return append(prefix, addr[:]...)
}

//RametronenterpriseStakeReturnsKeyPrefix returns the prefix of the rametronenterprise stake return key
func RametronenterpriseStakeReturnsKeyPrefix() common.Bytes {
	return common.Bytes("ls/rametronenterprisesrk/")
}

//RametronenterpriseStakeReturnsKey returns the Rametronenterprise stake return key for the given height
func RametronenterpriseStakeReturnsKey(height uint64) common.Bytes {
	heightStr := strconv.FormatUint(height, 10)
	return common.Bytes(string(RametronenterpriseStakeReturnsKeyPrefix()) + heightStr)
}

func RametronenterprisesTotalActiveStakeKey() common.Bytes {
	return common.Bytes("ls/rametronenterprisetas")
}
