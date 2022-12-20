package state

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"

	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/common/util"
	"github.com/pandoprojects/pando/core"
	"github.com/pandoprojects/pando/crypto/bls"
	"github.com/pandoprojects/pando/ledger/types"
)

const rametronenterprisepRewardN = 800 // Reward receiver sampling params

//
// ------- RametronenterprisePool ------- //
//

type RametronenterprisePool struct {
	readOnly bool
	sv       *StoreView
}

// NewRametronenterprisePool creates a new instance of RametronenterprisePool.
func NewRametronenterprisePool(sv *StoreView, readOnly bool) *RametronenterprisePool {
	return &RametronenterprisePool{
		readOnly: readOnly,
		sv:       sv,
	}
}

// Contains checks if given address is in the pool.
func (rametronenterprisep *RametronenterprisePool) RandomRewardWeight(block common.Hash, rametronenterpriseAddr common.Address) int {
	rametronenterprise := rametronenterprisep.Get(rametronenterpriseAddr)
	if rametronenterprise == nil {
		logger.Debugf("rametronenterprise random reward weight: address = %v, block = %v, weight = 0, not staked yet", rametronenterpriseAddr, block.Hex())
		return 0
	}
	totalStake := rametronenterprisep.sv.GetTotalRametronenterpriseStake()
	stake := rametronenterprise.TotalStake()

	seed := make([]byte, common.HashLength+common.AddressLength)
	copy(seed, block.Bytes())
	copy(seed[common.HashLength:], rametronenterpriseAddr[:])
	weight := sampleRametronenterpriseWeight(util.NewHashRand(seed), stake, totalStake)

	//logger.Debugf("rametronenterprise random reward weight: address = %v, block = %v, weight = %v, stake = %v, totalStake = %v", rametronenterpriseAddr, block.Hex(), weight, stake, totalStake)

	return weight
}

//
// The following sampling algorithm is based on Algorand's crypto sortition to randomly sample Rametronenterprises.
// Denote the expected TOTAL number of selected "stake units" by n, the stake of the Rametronenterprise stake by S.
// We essentially flip a biased coin (with head probability p) floor(S/S_min) times. And count the
// number of heads as the number of selected "stake units" of this Rametronenterprise.
//
// The head probability p = min(1.0, a * n * S_min / S_total), where a = (S/S_min) / floor(S/S_min).
// The factor a is to compensate the cases where the stake S is not a multiple of S_min. It can be
// proved that if a user split the stakes onto multiple nodes, the expected return won't changes, the
// variance changes a bit but shouldn't be too big.
//
func sampleRametronenterpriseWeight(reader io.Reader, stake *big.Int, totalStake *big.Int) int {
	if stake.Cmp(big.NewInt(0)) == 0 || totalStake.Cmp(big.NewInt(0)) == 0 {
		// could happen when we sample an Rametronenterprise whose stakes are all withdrawn, e.g. when
		// validating the votes from an Rametronenterprise with all stakes withdrawn
		return 0
	}

	if totalStake.Cmp(big.NewInt(0)) < 0 {
		logger.Panicf("Negative total stake: %v", totalStake)
	}

	b := new(big.Int).Div(stake, core.MinRametronenterpriseStakeDeposit)

	base := new(big.Int).SetUint64(1e18)

	p := new(big.Int).Mul(base, big.NewInt(rametronenterprisepRewardN))
	p.Mul(p, stake)
	p.Div(p, totalStake)
	p.Div(p, b)

	weight := 0
	for i := 0; i < int(b.Int64()); i++ {
		r, err := crand.Int(reader, base)
		if err != nil {
			log.Panicf("Failed to generate random number: %v", err)
		}
		if r.Cmp(p) < 0 {
			weight++
		}

		//logger.Debugf("rametronenterprise sampling: p = %v, r = %v, base = %v, weight = %v, stake = %v, totalStake = %v", p, r, base, weight, stake, totalStake)
	}

	return weight
}

// Contains checks if given address is in the pool.
func (rametronenterprisep *RametronenterprisePool) Contains(rametronenterpriseAddr common.Address) bool {
	return (rametronenterprisep.Get(rametronenterpriseAddr) != nil)
}

// GetPubKeys returns BLS pubkeys of given addresses.
func (rametronenterprisep *RametronenterprisePool) GetPubKeys(rametronenterpriseAddrs []common.Address) []*bls.PublicKey {
	ret := []*bls.PublicKey{}
	for _, addr := range rametronenterpriseAddrs {
		rametronenterprise := rametronenterprisep.Get(addr)
		if rametronenterprise == nil {
			return nil
		}
		ret = append(ret, rametronenterprise.Pubkey)
	}
	return ret
}

// Get returns the Rametronenterprise if exists, nil otherwise
func (rametronenterprisep *RametronenterprisePool) Get(rametronenterpriseAddr common.Address) *core.Rametronenterprise {
	rametronenterpriseKey := RametronenterpriseKey(rametronenterpriseAddr)
	data := rametronenterprisep.sv.Get(rametronenterpriseKey)
	if data == nil || len(data) == 0 {
		return nil
	}

	rametronenterprise := &core.Rametronenterprise{}
	err := types.FromBytes(data, rametronenterprise)
	if err != nil {
		log.Panicf("RametronenterprisePool.Get: Error reading rametronenterprise %X, error: %v",
			data, err.Error())
	}

	return rametronenterprise
}

// Upsert update or insert an rametronenterprise
func (rametronenterprisep *RametronenterprisePool) Upsert(rametronenterprise *core.Rametronenterprise) {
	if rametronenterprisep.readOnly {
		log.Panicf("RametronenterprisePool.Upsert: the pool is read-only")
	}

	rametronenterpriseKey := RametronenterpriseKey(rametronenterprise.Holder)
	data, err := types.ToBytes(rametronenterprise)
	if err != nil {
		log.Panicf("RametronenterprisePool.Upsert: Error serializing rametronenterprise %X, error: %v",
			data, err.Error())
	}
	rametronenterprisep.sv.Set(rametronenterpriseKey, data)
}

// Remove deletes the rametronenterprise from the pool
func (rametronenterprisep *RametronenterprisePool) Remove(rametronenterprise *core.Rametronenterprise) {
	if rametronenterprisep.readOnly {
		log.Panicf("RametronenterprisePool.Upsert: the pool is read-only")
	}

	rametronenterpriseKey := RametronenterpriseKey(rametronenterprise.Holder)
	rametronenterprisep.sv.Delete(rametronenterpriseKey)
}

func (rametronenterprisep *RametronenterprisePool) GetAll(withstake bool) []*core.Rametronenterprise {
	prefix := RametronenterpriseKeyPrefix()

	rametronenterpriseList := []*core.Rametronenterprise{}
	cb := func(k, v common.Bytes) bool {
		rametronenterprise := &core.Rametronenterprise{}
		err := types.FromBytes(v, rametronenterprise)
		if err != nil {
			log.Panicf("RametronenterprisePool.GetAll: Error reading rametronenterprise %X, error: %v",
				v, err.Error())
		}
		if withstake {
			hasStake := false
			for _, stake := range rametronenterprise.Stakes {
				if !stake.Withdrawn {
					hasStake = true
					break
				}
			}
			if !hasStake {
				return true // Skip if rametronenterprise dons't have non-withdrawn stake
			}
		}
		rametronenterpriseList = append(rametronenterpriseList, rametronenterprise)
		return true
	}

	rametronenterprisep.sv.Traverse(prefix, cb)

	return rametronenterpriseList
}

func (rametronenterprisep *RametronenterprisePool) DepositStake(source common.Address, holder common.Address, amount *big.Int, pubkey *bls.PublicKey, blockHeight uint64) (err error) {
	if rametronenterprisep.readOnly {
		log.Panicf("RametronenterprisePool.DepositStake: the pool is read-only")
	}

	
	// minRametronenterpriseStake := core.MinRametronenterpriseStakeDeposit

	// maxRametronenterpriseStake := core.MaxRametronenterpriseStakeDeposit
	// if amount.Cmp(minRametronenterpriseStake) < 0 {
	// 	return fmt.Errorf("rametronenterprise staking amount below the lower limit: %v", amount)
	// }
	// if amount.Cmp(maxRametronenterpriseStake) > 0 {
	// 	return fmt.Errorf("rametronenterprise staking amount above the upper limit: %v", amount)
	// }

	rametronenterprise := rametronenterprisep.Get(holder)
	if rametronenterprise == nil {
		rametronenterprise = core.NewRametronenterprise(
			core.NewStakeHolder(holder, []*core.Stake{core.NewStake(source, amount)}),
			pubkey)
	} else {
		if rametronenterprise.Holder != holder {
			log.Panicf("RametronenterprisePool.DepositStake: holder mismatch, rametronenterprise.Holder = %v, holder = %v",
				rametronenterprise.Holder.Hex(), holder.Hex())
		}
		// currentStake := rametronenterprise.TotalStake()
		// expectedStake := big.NewInt(0).Add(currentStake, amount)
		// if expectedStake.Cmp(maxRametronenterpriseStake) > 0 {
		// 	return fmt.Errorf("rametronenterprise stake would exceed the cap: %v", expectedStake)
		// }
		err = rametronenterprise.DepositStake(source, amount)
		if err != nil {
			return err
		}
	}

	rametronenterprisep.Upsert(rametronenterprise)

	// Update total rametronenterprisep stake
	totalStake := rametronenterprisep.sv.GetTotalRametronenterpriseStake()
	totalStake.Add(totalStake, amount)
	rametronenterprisep.sv.SetTotalRametronenterpriseStake(totalStake)

	return nil
}

func (rametronenterprisep *RametronenterprisePool) WithdrawStake(source common.Address, holder common.Address, currentHeight uint64) (*core.Stake, error) {
	if rametronenterprisep.readOnly {
		log.Panicf("RametronenterprisePool.WithdrawStake: the pool is read-only")
	}

	var withdrawnStake *core.Stake
	var err error

	rametronenterprise := rametronenterprisep.Get(holder)
	if rametronenterprise == nil {
		return nil, fmt.Errorf("No matched stake holder address found: %v", holder)
	}

	if rametronenterprise.Holder != holder {
		log.Panicf("RametronenterprisePool.DepositStake: holder mismatch, rametronenterprise.Holder = %v, holder = %v",
			rametronenterprise.Holder.Hex(), holder.Hex())
	}

	withdrawnStake, err = rametronenterprise.WithdrawStake(source, currentHeight)
	if err != nil {
		return nil, err
	}

	rametronenterprisep.Upsert(rametronenterprise)

	// Update total rametronenterprisep stake
	totalStake := rametronenterprisep.sv.GetTotalRametronenterpriseStake()
	totalStake.Sub(totalStake, withdrawnStake.Amount)
	rametronenterprisep.sv.SetTotalRametronenterpriseStake(totalStake)

	return withdrawnStake, nil
}

func (rametronenterprisep *RametronenterprisePool) ReturnStake(currentHeight uint64, holder common.Address, returnedStake core.Stake) error {
	rametronenterprise := rametronenterprisep.Get(holder)
	if rametronenterprise == nil {
		return fmt.Errorf("No matched stake holder address found: %v", holder)
	}

	sourceAddress := returnedStake.Source
	numStakes := len(rametronenterprise.Stakes)

	// need to iterate in the reverse order, since we may delete elemements from the slice while iterating through it
	for sidx := numStakes - 1; sidx >= 0; sidx-- {
		stake := rametronenterprise.Stakes[sidx]

		if stake.Source == sourceAddress {
			if stake.Withdrawn == false || stake.ReturnHeight != currentHeight {
				log.Panicf("Returned stake mismatch: rametronenterpriseAddr = %v, sourceAddr = %v, currentHeight = %v, stake.Withdrawn = %v, stake.ReturnHeight = %v",
					holder, sourceAddress, currentHeight, stake.Withdrawn, stake.ReturnHeight)
			}

			logger.Infof("Stake to be returned: source = %v, amount = %v", stake.Source, stake.Amount)
			_, err := rametronenterprise.ReturnStake(sourceAddress, currentHeight)
			if err != nil {
				return err
			}

			if len(rametronenterprise.Stakes) == 0 { // the candidate's stake becomes zero, no need to keep track of the candidate anymore
				rametronenterprisep.Remove(rametronenterprise)
			} else {
				rametronenterprisep.Upsert(rametronenterprise)
			}

			break // only one stake to be returned
		}
	}

	return nil
}
