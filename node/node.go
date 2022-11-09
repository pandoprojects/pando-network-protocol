package node

import (
	"context"
	"log"
	"reflect"
	"sync"

	"github.com/spf13/viper"
	"github.com/pandoprojects/pando/blockchain"
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/consensus"
	"github.com/pandoprojects/pando/core"
	"github.com/pandoprojects/pando/crypto"
	dp "github.com/pandoprojects/pando/dispatcher"
	ld "github.com/pandoprojects/pando/ledger"
	mp "github.com/pandoprojects/pando/mempool"
	"github.com/pandoprojects/pando/netsync"
	"github.com/pandoprojects/pando/p2p"
	"github.com/pandoprojects/pando/p2pl"
	rp "github.com/pandoprojects/pando/report"
	"github.com/pandoprojects/pando/rpc"
	"github.com/pandoprojects/pando/snapshot"
	"github.com/pandoprojects/pando/store"
	"github.com/pandoprojects/pando/store/database"
	"github.com/pandoprojects/pando/store/kvstore"
	"github.com/pandoprojects/pando/store/rollingdb"
)

type Node struct {
	Store            store.Store
	Chain            *blockchain.Chain
	Consensus        *consensus.ConsensusEngine
	ValidatorManager core.ValidatorManager
	SyncManager      *netsync.SyncManager
	Dispatcher       *dp.Dispatcher
	Ledger           core.Ledger
	Mempool          *mp.Mempool
	RPC              *rpc.PandoRPCServer
	reporter         *rp.Reporter

	// Life cycle
	wg      *sync.WaitGroup
	quit    chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	stopped bool
}

type Params struct {
	ChainID             string
	PrivateKey          *crypto.PrivateKey
	Root                *core.Block
	NetworkOld          p2p.Network
	Network             p2pl.Network
	DB                  database.Database
	RollingDB           *rollingdb.RollingDB
	SnapshotPath        string
	ChainImportDirPath  string
	ChainCorrectionPath string
}

func NewNode(params *Params) *Node {
	store := kvstore.NewKVStore(params.DB)
	chain := blockchain.NewChain(params.ChainID, store, params.Root)
	params.RollingDB.SetChain(chain)

	validatorManager := consensus.NewRotatingValidatorManager()
	dispatcher := dp.NewDispatcher(params.NetworkOld, params.Network)

	consensus := consensus.NewConsensusEngine(params.PrivateKey, store, chain, dispatcher, validatorManager)
	reporter := rp.NewReporter(dispatcher, consensus, chain)

	// TODO: check if this is a guardian node
	syncMgr := netsync.NewSyncManager(chain, consensus, params.NetworkOld, params.Network, dispatcher, consensus, reporter)
	mempool := mp.CreateMempool(dispatcher, consensus)
	ledger := ld.NewLedger(params.ChainID, params.RollingDB, params.RollingDB, chain, consensus, validatorManager, mempool)

	validatorManager.SetConsensusEngine(consensus)
	consensus.SetLedger(ledger)
	mempool.SetLedger(ledger)
	txMsgHandler := mp.CreateMempoolMessageHandler(mempool)

	if !reflect.ValueOf(params.Network).IsNil() {
		params.Network.RegisterMessageHandler(txMsgHandler)
	}
	if !reflect.ValueOf(params.NetworkOld).IsNil() {
		params.NetworkOld.RegisterMessageHandler(txMsgHandler)
	}

	currentHeight := consensus.GetLastFinalizedBlock().Height
	if currentHeight <= params.Root.Height {
		snapshotPath := params.SnapshotPath
		chainImportDirPath := params.ChainImportDirPath
		chainCorrectionPath := params.ChainCorrectionPath
		var lastCC *core.ExtendedBlock
		var err error
		if _, lastCC, err = snapshot.ImportSnapshot(snapshotPath, chainImportDirPath, chainCorrectionPath, chain, params.DB, ledger); err != nil {
			log.Fatalf("Failed to load snapshot: %v, err: %v", snapshotPath, err)
		}
		if lastCC != nil {
			state := consensus.State()
			state.SetLastFinalizedBlock(lastCC)
			state.SetHighestCCBlock(lastCC)
			state.SetLastVote(core.Vote{})
			state.SetLastProposal(core.Proposal{})
		}
	}

	node := &Node{
		Store:            store,
		Chain:            chain,
		Consensus:        consensus,
		ValidatorManager: validatorManager,
		SyncManager:      syncMgr,
		Dispatcher:       dispatcher,
		Ledger:           ledger,
		Mempool:          mempool,
		reporter:         reporter,
	}

	if viper.GetBool(common.CfgRPCEnabled) {
		node.RPC = rpc.NewPandoRPCServer(mempool, ledger, dispatcher, chain, consensus)
	}
	return node
}

// Start starts sub components and kick off the main loop.
func (n *Node) Start(ctx context.Context) {
	c, cancel := context.WithCancel(ctx)
	n.ctx = c
	n.cancel = cancel

	n.Consensus.Start(n.ctx)
	n.SyncManager.Start(n.ctx)
	n.Dispatcher.Start(n.ctx)
	n.Mempool.Start(n.ctx)
	n.reporter.Start(n.ctx)

	if viper.GetBool(common.CfgRPCEnabled) {
		n.RPC.Start(n.ctx)
	}
}

// Stop notifies all sub components to stop without blocking.
func (n *Node) Stop() {
	n.cancel()
}

// Wait blocks until all sub components stop.
func (n *Node) Wait() {
	n.Consensus.Wait()
	n.SyncManager.Wait()
	if n.RPC != nil {
		n.RPC.Wait()
	}
}
