package core

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/common/result"
	"github.com/pandoprojects/pando/crypto/bls"
	"github.com/pandoprojects/pando/rlp"
)

//
// ------- RametronenterpriseVote ------- //
//

// RametronenterpriseVote represents the vote for a block from an rametronenterprise.
type RametronenterpriseVote struct {
	Block     common.Hash    // Hash of the block.
	Height    uint64         // Height of the block, just for reference
	Address   common.Address // Address of the rametronenterprise.
	Signature *bls.Signature // Aggregated signiature.
	Timestamp *big.Int       // Unix timestamp when the RametronenterpriseVote was created, just for reference
}

type RametronenterpriseBlsSigMsg struct {
	Block common.Hash
}

func NewRametronenterpriseVote(block common.Hash, blockHeight uint64, rametronenterpriseAddr common.Address, signature *bls.Signature) *RametronenterpriseVote {
	return &RametronenterpriseVote{
		Block:     block,
		Height:    blockHeight,
		Address:   rametronenterpriseAddr,
		Signature: signature,
		Timestamp: big.NewInt(time.Now().Unix()),
	}
}

// signBytes returns the bytes to be signed.
func (e *RametronenterpriseVote) signBytes() common.Bytes {
	tmp := &RametronenterpriseBlsSigMsg{
		Block: e.Block,
	}
	b, _ := rlp.EncodeToBytes(tmp)
	return b
}

// Validate verifies the vote.
func (e *RametronenterpriseVote) Validate(rametronenterpriseBLSPubkey *bls.PublicKey) result.Result {
	if e.Signature == nil {
		return result.Error("signature cannot be nil")
	}
	if !e.Signature.Verify(e.signBytes(), rametronenterpriseBLSPubkey) {
		return result.Error("rametronenterprise vote signature validation failed")
	}
	return result.OK
}

func (e *RametronenterpriseVote) String() string {
	return fmt.Sprintf("RametronenterpriseVote{Block: %s, Height: %v, Address: %v, Signature: %v, CreationTimestamp: %v}",
		e.Block.Hex(), e.Height, e.Address, e.signBytes(), e.Timestamp)
}

//
// ------- AggregatedRametronenterpriseVotes ------- //
//

// AggregatedRametronenterpriseVotes represents the aggregated rametronenterprise votes on a block.
type AggregatedRametronenterpriseVotes struct {
	Block      common.Hash      // Hash of the block.
	Multiplies []uint32         // Multiplies of each signer.
	Addresses  []common.Address // Addresses of each signer
	Signature  *bls.Signature   // Aggregated signature.
}

func NewAggregatedRametronenterpriseVotes(block common.Hash) *AggregatedRametronenterpriseVotes {
	return &AggregatedRametronenterpriseVotes{
		Block:     block,
		Signature: bls.NewAggregateSignature(),
	}
}

func (a *AggregatedRametronenterpriseVotes) String() string {
	return fmt.Sprintf("AggregatedRametronenterpriseVotes{Block: %s, Addresses: %v, Multiplies: %v}", a.Block.Hex(), a.Addresses, a.Multiplies)
}

// signBytes returns the bytes to be signed.
func (a *AggregatedRametronenterpriseVotes) signBytes() common.Bytes {
	// tmp := &AggregatedRametronenterpriseVotes{
	// 	Block: a.Block,
	// }
	tmp := &RametronenterpriseBlsSigMsg{
		Block: a.Block,
	}
	b, _ := rlp.EncodeToBytes(tmp)
	return b
}

// // Sign adds signer's signature. Returns false if signer has already signed.
// func (a *AggregatedRametronenterpriseVotes) Sign(key *bls.SecretKey, signerIdx int) bool {
// 	if a.Multiplies[signerIdx] > 0 {
// 		// Already signed, do nothing.
// 		return false
// 	}

// 	a.Multiplies[signerIdx] = 1
// 	a.Signature.Aggregate(key.Sign(a.signBytes()))
// 	return true
// }

// Merge creates a new aggregation that combines two vote sets. Returns nil, nil if input vote
// is a subset of current vote.
func (a *AggregatedRametronenterpriseVotes) Merge(b *AggregatedRametronenterpriseVotes) (*AggregatedRametronenterpriseVotes, error) {
	if a.Block != b.Block {
		return nil, errors.New("Cannot merge incompatible votes")
	}
	newMultiplies := []uint32{}
	newAddresses := []common.Address{}

	isSubset := true
	i := 0
	j := 0
	for i < len(a.Addresses) && j < len(b.Addresses) {
		cmp := bytes.Compare(a.Addresses[i].Bytes(), b.Addresses[j].Bytes())
		if cmp == 0 {
			sum := a.Multiplies[i] + b.Multiplies[j]
			if sum < a.Multiplies[i] || sum < b.Multiplies[j] {
				return nil, errors.New("Signiature multipliers overflowed")
			}
			newAddresses = append(newAddresses, a.Addresses[i])
			newMultiplies = append(newMultiplies, sum)
			i++
			j++
		} else if cmp < 0 {
			newMultiplies = append(newMultiplies, a.Multiplies[i])
			newAddresses = append(newAddresses, a.Addresses[i])
			i++
			// Here we don't mark isSubset to false
		} else {
			newMultiplies = append(newMultiplies, b.Multiplies[j])
			newAddresses = append(newAddresses, b.Addresses[j])
			j++
			isSubset = false
		}
	}
	if i < len(a.Addresses) {
		newMultiplies = append(newMultiplies, a.Multiplies[i:]...)
		newAddresses = append(newAddresses, a.Addresses[i:]...)
	}
	if j < len(b.Addresses) {
		newMultiplies = append(newMultiplies, b.Multiplies[j:]...)
		newAddresses = append(newAddresses, b.Addresses[j:]...)
		isSubset = false
	}

	if isSubset {
		// The other vote is a subset of current vote
		return nil, nil
	}
	newSig := a.Signature.Copy()
	newSig.Aggregate(b.Signature)
	return &AggregatedRametronenterpriseVotes{
		Block:      a.Block,
		Multiplies: newMultiplies,
		Addresses:  newAddresses,
		Signature:  newSig,
	}, nil
}

// Abs returns the number of voted rametronenterprises in the vote
func (a *AggregatedRametronenterpriseVotes) Abs() int {
	ret := 0
	for i := 0; i < len(a.Multiplies); i++ {
		if a.Multiplies[i] != 0 {
			ret += 1
		}
	}
	return ret
}

// Pick selects better vote from two votes.
func (a *AggregatedRametronenterpriseVotes) Pick(b *AggregatedRametronenterpriseVotes) (*AggregatedRametronenterpriseVotes, error) {
	if a.Block != b.Block {
		return nil, errors.New("Cannot compare incompatible votes")
	}
	if b.Abs() > a.Abs() {
		return b, nil
	}
	return a, nil
}

// Validate performs basic validation of the voteset.
func (a *AggregatedRametronenterpriseVotes) Validate(rametronenterprisep RametronenterprisePool) result.Result {
	if rametronenterprisep == nil {
		return result.Error("empty rametronenterprisep")
	}
	if a.Signature == nil {
		return result.Error("signature cannot be nil")
	}
	if len(a.Addresses) == 0 {
		return result.Error("aggregated vote is empty")
	}
	if len(a.Addresses) != len(a.Multiplies) {
		return result.Error("aggregate vote lengths are inconsisent")
	}
	for i := 0; i < len(a.Addresses)-1; i++ {
		if bytes.Compare(a.Addresses[i].Bytes(), a.Addresses[i+1].Bytes()) >= 0 {
			return result.Error("aggregate vote addresses must be sorted")
		}
	}
	for _, addr := range a.Addresses {
		weight := rametronenterprisep.RandomRewardWeight(a.Block, addr)
		if weight == 0 {
			return result.Error("aggregate vote contains rametronenterprise that are not selected for checkpoint reward")
		}
	}
	pubkeys := rametronenterprisep.GetPubKeys(a.Addresses)
	aggPubkey := bls.AggregatePublicKeysVec(pubkeys, a.Multiplies)
	if !a.Signature.Verify(a.signBytes(), aggPubkey) {
		return result.Error("signature verification failed rm")
	}

	return result.OK
}

// Copy clones the aggregated votes
func (a *AggregatedRametronenterpriseVotes) Copy() *AggregatedRametronenterpriseVotes {
	clone := &AggregatedRametronenterpriseVotes{
		Block: a.Block,
	}
	if a.Multiplies != nil {
		clone.Multiplies = make([]uint32, len(a.Multiplies))
		copy(clone.Multiplies, a.Multiplies)
	}
	if a.Addresses != nil {
		clone.Addresses = make([]common.Address, len(a.Addresses))
		copy(clone.Addresses, a.Addresses)
	}
	if a.Signature != nil {
		clone.Signature = a.Signature.Copy()
	}

	return clone
}

var (
	MinRametronenterpriseStakeDeposit *big.Int
	MinRametronproStakeDeposit *big.Int
	MinRametronliteStakeDeposit *big.Int
	MinRametronmobileStakeDeposit *big.Int
	// MaxRametronenterpriseStakeDeposit *big.Int
)

func init() {
	// Each rametronenterprise stake deposit needs to be at least 10,000 PTX
	MinRametronenterpriseStakeDeposit = new(big.Int).Mul(new(big.Int).SetUint64(35000), new(big.Int).SetUint64(1e18))
	MinRametronproStakeDeposit = new(big.Int).Mul(new(big.Int).SetUint64(10000), new(big.Int).SetUint64(1e18))
	MinRametronliteStakeDeposit = new(big.Int).Mul(new(big.Int).SetUint64(1000), new(big.Int).SetUint64(1e18))
	MinRametronmobileStakeDeposit = new(big.Int).Mul(new(big.Int).SetUint64(250), new(big.Int).SetUint64(1e18))

	// Each rametronenterprise stake deposit should not exceed 500,000 PTX
	// MaxRametronenterpriseStakeDeposit = new(big.Int).Mul(new(big.Int).SetUint64(500000), new(big.Int).SetUint64(1e18))
}

//
// ------- Rametronenterprise ------- //
//

type Rametronenterprise struct {
	*StakeHolder
	Pubkey *bls.PublicKey `json:"-"`
}

func NewRametronenterprise(stakeHolder *StakeHolder, pubkey *bls.PublicKey) *Rametronenterprise {
	return &Rametronenterprise{
		StakeHolder: stakeHolder,
		Pubkey:      pubkey,
	}
}

func (rametronenterprise *Rametronenterprise) String() string {
	return fmt.Sprintf("{holder: %v, pubkey: %v, stakes :%v}", rametronenterprise.Holder, rametronenterprise.Pubkey.String(), rametronenterprise.Stakes)
}

func (rametronenterprise *Rametronenterprise) DepositStake(source common.Address, amount *big.Int) error {
	return rametronenterprise.StakeHolder.depositStake(source, amount)
}

func (rametronenterprise *Rametronenterprise) WithdrawStake(source common.Address, currentHeight uint64) (*Stake, error) {
	return rametronenterprise.StakeHolder.withdrawStake(source, currentHeight)
}

func (rametronenterprise *Rametronenterprise) ReturnStake(source common.Address, currentHeight uint64) (*Stake, error) {
	return rametronenterprise.StakeHolder.returnStake(source, currentHeight)
}

//
// ------- RametronenterprisePool ------- //
//

type RametronenterprisePool interface {
	Contains(rametronenterpriseAddr common.Address) bool
	GetPubKeys(rametronenterpriseAddrs []common.Address) []*bls.PublicKey
	Get(rametronenterpriseAddr common.Address) *Rametronenterprise
	Upsert(rametronenterprise *Rametronenterprise)
	GetAll(withstake bool) []*Rametronenterprise
	DepositStake(source common.Address, holder common.Address, amount *big.Int, pubkey *bls.PublicKey, blockHeight uint64) (err error)
	WithdrawStake(source common.Address, holder common.Address, currentHeight uint64) (*Stake, error)
	RandomRewardWeight(block common.Hash, rametronenterpriseAddr common.Address) int
}
