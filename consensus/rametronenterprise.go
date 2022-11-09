package consensus

import (
	"context"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/common/util"
	"github.com/pandoprojects/pando/core"
	"github.com/pandoprojects/pando/crypto/bls"
)

const (
	maxRametronenterpriseLogNeighbors    uint32 = 3 // Estimated number of neighbors during gossip = 2**3 = 8
	maxRametronenterpriseRound                  = 20
	sampleResultCacheSize        = 1000000
)

type RametronenterpriseEngine struct {
	logger *log.Entry

	engine  *ConsensusEngine
	privKey *bls.SecretKey

	voteBookkeeper *RametronenterpriseVoteBookkeeper

	// State for current voting
	block           common.Hash
	round           uint32
	currVote        *core.AggregatedRametronenterpriseVotes
	nextVote        *core.AggregatedRametronenterpriseVotes
	rametronenterprisep            core.RametronenterprisePool
	rametronenterpriseSampleResult *lru.Cache

	evIncoming  chan *core.RametronenterpriseVote
	aevIncoming chan *core.AggregatedRametronenterpriseVotes
	mu          *sync.Mutex
}

func NewRametronenterpriseEngine(c *ConsensusEngine, privateKey *bls.SecretKey) *RametronenterpriseEngine {
	return &RametronenterpriseEngine{
		logger:  util.GetLoggerForModule("rametronenterprise"),
		engine:  c,
		privKey: privateKey,

		voteBookkeeper: CreateRametronenterpriseVoteBookkeeper(DefaultMaxNumVotesCached),

		evIncoming:  make(chan *core.RametronenterpriseVote, viper.GetInt(common.CfgConsensusRametronenterpriseVoteQueueSize)),
		aevIncoming: make(chan *core.AggregatedRametronenterpriseVotes, viper.GetInt(common.CfgConsensusRametronenterpriseVoteQueueSize)),
		mu:          &sync.Mutex{},
	}
}

func (e *RametronenterpriseEngine) StartNewBlock(block common.Hash) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.block = block
	e.nextVote = nil
	e.currVote = nil
	e.round = 1

	rametronenterprisep, err := e.engine.GetLedger().GetRametronenterprisePoolOfLastCheckpoint(block)
	if err != nil {
		// Should not happen
		e.logger.Panic(err)
	}
	e.rametronenterprisep = rametronenterprisep
	e.rametronenterpriseSampleResult, err = lru.New(sampleResultCacheSize)
	if err != nil {
		e.logger.Panic(err)
	}

	e.logger.WithFields(log.Fields{
		"block": block.Hex(),
	}).Debug("Starting new block")

	if viper.GetBool(common.CfgDebugLogSelectedRametronenterprisePs) {
		count := 0
		total := 0
		for _, rametronenterprise := range rametronenterprisep.GetAll(true) {
			total++
			if rametronenterprisep.RandomRewardWeight(block, rametronenterprise.Holder) > 0 {
				count++
				logger.Debugf("selected Rametronenterprise: %v, block: %v", rametronenterprise.Holder, block.Hex())
			}
		}

		e.logger.WithFields(log.Fields{
			"block": block.Hex(),
			"count": count,
			"total": total,
		}).Debug("Selected Rametronenterprises")
	}
}

func (e *RametronenterpriseEngine) StartNewRound() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.round < maxRametronenterpriseRound {
		e.round++
		if e.nextVote != nil {
			e.currVote = e.nextVote.Copy()
		}
	}
}

func (e *RametronenterpriseEngine) GetVoteToBroadcast() *core.AggregatedRametronenterpriseVotes {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.currVote
}

func (e *RametronenterpriseEngine) GetBestVote() *core.AggregatedRametronenterpriseVotes {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.nextVote
}

func (e *RametronenterpriseEngine) Start(ctx context.Context) {
	go e.mainLoop(ctx)
}

func (e *RametronenterpriseEngine) mainLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-e.evIncoming:
			if ok {
				e.processVote(ev)
			}
		case aev, ok := <-e.aevIncoming:
			if ok {
				e.processAggregatedVote(aev)
			}
		}
	}
}

func (e *RametronenterpriseEngine) processVote(vote *core.RametronenterpriseVote) {
	e.mu.Lock()
	defer e.mu.Unlock()

	logger.Debugf("Process rametronenterprise vote {%v : %v}", vote.Address, vote.Block.Hex())

	if !e.validateVote(vote) {
		return
	}

	logger.Debugf("Validated rametronenterprise vote {%v : %v}", vote.Address, vote.Block.Hex())

	aggregatedVote, err := e.convertVote(vote)
	if err != nil {
		logger.Warnf("Discard vote from rametronenterprise %v, reason: %v", vote.Address, err)
		return
	}

	logger.Debugf("Converted rametronenterprise vote to aggregated vote {%v : %v}", vote.Address, vote.Block.Hex())

	e.aevIncoming <- aggregatedVote
}

// convertVote converts an RametronenterpriseVote into an AggregatedRametronenterpriseVotes
func (e *RametronenterpriseEngine) convertVote(ev *core.RametronenterpriseVote) (*core.AggregatedRametronenterpriseVotes, error) {
	rametronenterprisev := core.NewAggregatedRametronenterpriseVotes(ev.Block)
	rametronenterprisev.Multiplies = []uint32{1}
	rametronenterprisev.Addresses = []common.Address{ev.Address}
	rametronenterprisev.Signature = ev.Signature

	logger.Infof("converted rametronenterprise vote for block %v from rametronenterprise %v to an aggregated vote", ev.Block.Hex(), ev.Address)

	return rametronenterprisev, nil
}

func (e *RametronenterpriseEngine) processAggregatedVote(vote *core.AggregatedRametronenterpriseVotes) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.validateAggregatedVote(vote) {
		return
	}

	if e.nextVote == nil {
		e.nextVote = vote
		return
	}

	var candidate *core.AggregatedRametronenterpriseVotes
	var err error

	candidate, err = e.nextVote.Merge(vote)
	if err != nil {
		e.logger.WithFields(log.Fields{
			"e.block":               e.block.Hex(),
			"e.round":               e.round,
			"vote.block":            vote.Block.Hex(),
			"vote.Mutiplies":        vote.Multiplies,
			"e.nextVote.Multiplies": e.nextVote.Multiplies,
			"e.nextVote.Block":      e.nextVote.Block.Hex(),
			"error":                 err.Error(),
		}).Debug("Failed to merge aggregated rametronenterprise vote")
	}
	if candidate == nil {
		// Incoming vote is subset of the current nextVote.
		e.logger.WithFields(log.Fields{
			"vote.block":     vote.Block.Hex(),
			"vote.Mutiplies": vote.Multiplies,
		}).Debug("Skipping aggregated rametronenterprise vote: no new index")
		return
	}

	if !e.checkMultipliesForRound(candidate, e.round+1) {
		e.logger.WithFields(log.Fields{
			"local.block":           e.block.Hex(),
			"local.round":           e.round,
			"vote.block":            vote.Block.Hex(),
			"vote.Mutiplies":        vote.Multiplies,
			"local.vote.Multiplies": e.nextVote.Multiplies,
		}).Debug("Skipping aggregated rametronenterprise vote: candidate vote overflows")
		return
	}

	e.nextVote = candidate

	e.logger.WithFields(log.Fields{
		"local.block":           e.block.Hex(),
		"local.round":           e.round,
		"local.vote.Multiplies": e.nextVote.Multiplies,
	}).Debug("New aggregated rametronenterprise vote")
}

func (e *RametronenterpriseEngine) HandleVote(vote *core.RametronenterpriseVote) {
	if e.voteBookkeeper.HasSeen(vote) {
		//logger.Debugf("Received rametronenterprise vote {%v : %v} earlier, safely ignore", vote.Address, vote.Block.Hex())
		return
	}
	e.voteBookkeeper.Record(vote)

	logger.Debugf("Received rametronenterprise vote {%v : %v} for the first time", vote.Address, vote.Block.Hex())

	select {
	case e.evIncoming <- vote:
		return
	default:
		e.logger.Debug("RametronenterpriseEngine queue is full, discarding rametronenterprise vote: %v", vote)
	}
}

func (e *RametronenterpriseEngine) HandleAggregatedVote(vote *core.AggregatedRametronenterpriseVotes) {
	select {
	case e.aevIncoming <- vote:
		return
	default:
		e.logger.Debug("RametronenterpriseEngine queue is full, discarding aggregated rametronenterprise vote: %v", vote)
	}
}

func (e *RametronenterpriseEngine) validateVote(vote *core.RametronenterpriseVote) (res bool) {
	if e.rametronenterprisep == nil {
		// e.logger.WithFields(log.Fields{
		// 	"local.block":  e.block.Hex(),
		// 	"local.round":  e.round,
		// 	"vote.block":   vote.Block.Hex(),
		// 	"vote.address": vote.Address,
		// }).Debug("The rametronenterprise pool is nil, cannot validate vote")
		return
	}

	if e.block.IsEmpty() {
		// e.logger.WithFields(log.Fields{
		// 	"local.block":  e.block.Hex(),
		// 	"local.round":  e.round,
		// 	"vote.block":   vote.Block.Hex(),
		// 	"vote.address": vote.Address,
		// }).Debug("Ignoring rametronenterprise vote: local not ready")
		return
	}
	if vote.Block != e.block {
		// e.logger.WithFields(log.Fields{
		// 	"local.block":  e.block.Hex(),
		// 	"local.round":  e.round,
		// 	"vote.block":   vote.Block.Hex(),
		// 	"vote.address": vote.Address,
		// }).Debug("Ignoring rametronenterprise vote: block hash does not match with local candidate")
		return
	}

	// Check if rametronenterprise is selected for this round
	ok := e.rametronenterpriseSampleResult.Contains(vote.Address)
	if !ok {
		weight := e.rametronenterprisep.RandomRewardWeight(vote.Block, vote.Address)
		if weight == 0 {
			e.rametronenterpriseSampleResult.Add(vote.Address, false)
		} else {
			e.rametronenterpriseSampleResult.Add(vote.Address, true)
		}
	}
	selected, _ := e.rametronenterpriseSampleResult.Get(vote.Address)
	if !selected.(bool) {
		// e.logger.WithFields(log.Fields{
		// 	"local.block":  e.block.Hex(),
		// 	"local.round":  e.round,
		// 	"vote.block":   vote.Block.Hex(),
		// 	"vote.address": vote.Address,
		// }).Debug("Ignoring rametronenterprise vote: not selected by random sampling")
		return
	}

	pubkeys := e.rametronenterprisep.GetPubKeys([]common.Address{vote.Address})
	if len(pubkeys) != 1 {
		e.logger.WithFields(log.Fields{
			"local.block":  e.block.Hex(),
			"local.round":  e.round,
			"vote.block":   vote.Block.Hex(),
			"vote.address": vote.Address,
		}).Debug("Ignoring rametronenterprise vote: failed to get pubkey")
	}
	if result := vote.Validate(pubkeys[0]); result.IsError() {
		e.logger.WithFields(log.Fields{
			"local.block":  e.block.Hex(),
			"local.round":  e.round,
			"vote.block":   vote.Block.Hex(),
			"vote.address": vote.Address,
			"result":       result.Message,
		}).Debug("Ignoring rametronenterprise vote: invalid signature")
		return
	}

	res = true
	return
}

func (e *RametronenterpriseEngine) validateAggregatedVote(vote *core.AggregatedRametronenterpriseVotes) (res bool) {
	if e.block.IsEmpty() {
		e.logger.WithFields(log.Fields{
			"local.block":    e.block.Hex(),
			"local.round":    e.round,
			"vote.block":     vote.Block.Hex(),
			"vote.Addresses": vote.Addresses,
			"vote.Mutiplies": vote.Multiplies,
		}).Debug("Ignoring aggregated rametronenterprise vote: local not ready")
		return
	}
	if vote.Block != e.block {
		e.logger.WithFields(log.Fields{
			"local.block":    e.block.Hex(),
			"local.round":    e.round,
			"vote.block":     vote.Block.Hex(),
			"vote.Addresses": vote.Addresses,
			"vote.Mutiplies": vote.Multiplies,
		}).Debug("Ignoring aggregated rametronenterprise vote: block hash does not match with local candidate")
		return
	}
	if !e.checkMultipliesForRound(vote, e.round) {
		e.logger.WithFields(log.Fields{
			"local.block":    e.block.Hex(),
			"local.round":    e.round,
			"vote.block":     vote.Block.Hex(),
			"vote.Addresses": vote.Addresses,
			"vote.Mutiplies": vote.Multiplies,
		}).Debug("Ignoring aggregated rametronenterprise vote: mutiplies exceed limit for round")
		return
	}
	if result := vote.Validate(e.rametronenterprisep); result.IsError() {
		e.logger.WithFields(log.Fields{
			"local.block":    e.block.Hex(),
			"local.round":    e.round,
			"vote.block":     vote.Block.Hex(),
			"vote.Addresses": vote.Addresses,
			"vote.Mutiplies": vote.Multiplies,
			"error":          result.Message,
		}).Debug("Ignoring aggregated rametronenterprise vote: invalid vote")
		return
	}

	res = true
	return
}

func (e *RametronenterpriseEngine) checkMultipliesForRound(vote *core.AggregatedRametronenterpriseVotes, k uint32) bool {
	// for _, m := range vote.Multiplies {
	// 	if m > g.maxMultiply(k) {
	// 		return false
	// 	}
	// }
	return true
}

func (e *RametronenterpriseEngine) maxMultiply(k uint32) uint32 {
	return 1 << (k * maxRametronenterpriseLogNeighbors)
}
