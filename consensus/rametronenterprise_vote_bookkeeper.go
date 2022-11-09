package consensus

import (
	"container/list"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	tcore "github.com/pandoprojects/pando/core"
	tcrypto "github.com/pandoprojects/pando/crypto"
)

const DefaultMaxNumVotesCached = uint(5000000)

const maxVoteLife = 20 * time.Minute // more than one checkpoint interval

//
// RametronenterpriseVoteBookkeeper keeps tracks of recently seen rametronenterprise votes
//
type RametronenterpriseVoteBookkeeper struct {
	mutex *sync.Mutex

	voteMap  map[string]*RametronenterpriseVoteRecord // map: vote hash -> record
	voteList list.List                 // FIFO list of vote hashes

	maxNumVotes uint
}

type RametronenterpriseVoteRecord struct {
	Hash      string
	Count     uint
	CreatedAt time.Time
}

func (r *RametronenterpriseVoteRecord) IsOutdated() bool {
	return time.Since(r.CreatedAt) > maxVoteLife
}

type TxStatus int

const (
	TxStatusPending TxStatus = iota
	TxStatusAbandoned
)

func CreateRametronenterpriseVoteBookkeeper(maxNumTxs uint) *RametronenterpriseVoteBookkeeper {
	return &RametronenterpriseVoteBookkeeper{
		mutex:       &sync.Mutex{},
		voteMap:     make(map[string]*RametronenterpriseVoteRecord),
		maxNumVotes: maxNumTxs,
	}
}

func (vb *RametronenterpriseVoteBookkeeper) reset() {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	vb.voteMap = make(map[string]*RametronenterpriseVoteRecord)
	vb.voteList.Init()
}

func (vb *RametronenterpriseVoteBookkeeper) ReceiveCount(vote *tcore.RametronenterpriseVote) uint {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()

	// Remove outdated Tx records
	vb.removeOutdatedVotesUnsafe()

	voteHash := getVoteHash(vote)
	voteRecord, exists := vb.voteMap[voteHash]
	if !exists || voteRecord == nil {
		return 0
	}

	return voteRecord.Count
}

func (vb *RametronenterpriseVoteBookkeeper) HasSeen(vote *tcore.RametronenterpriseVote) bool {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()

	// Remove outdated Tx records
	vb.removeOutdatedVotesUnsafe()

	voteHash := getVoteHash(vote)
	_, exists := vb.voteMap[voteHash]
	return exists
}

func (vb *RametronenterpriseVoteBookkeeper) removeOutdatedVotesUnsafe() {
	// Loop and remove all outdated Tx records
	for {
		el := vb.voteList.Front()
		if el == nil {
			return
		}
		voteRecord := el.Value.(*RametronenterpriseVoteRecord)
		if !voteRecord.IsOutdated() {
			return
		}

		if _, exists := vb.voteMap[voteRecord.Hash]; exists {
			delete(vb.voteMap, voteRecord.Hash)
		}
		vb.voteList.Remove(el)
	}
}

func (vb *RametronenterpriseVoteBookkeeper) Record(vote *tcore.RametronenterpriseVote) bool {
	vb.mutex.Lock()
	defer vb.mutex.Unlock()
	voteHash := getVoteHash(vote)

	// Remove outdated vote records
	vb.removeOutdatedVotesUnsafe()

	if existingVoteRecord, exists := vb.voteMap[voteHash]; exists {
		existingVoteRecord.Count += 1
		return true
	}

	if uint(vb.voteList.Len()) >= vb.maxNumVotes { // remove the oldest votes
		popped := vb.voteList.Front()
		poppedVoteHash := popped.Value.(*RametronenterpriseVoteRecord).Hash
		delete(vb.voteMap, poppedVoteHash)
		vb.voteList.Remove(popped)
	}

	record := &RametronenterpriseVoteRecord{
		Hash:      voteHash,
		Count:     0,
		CreatedAt: time.Now(),
	}
	vb.voteMap[voteHash] = record

	vb.voteList.PushBack(record)

	return true
}

func getVoteHash(vote *tcore.RametronenterpriseVote) string {
	voteStr := fmt.Sprintf("%v:%v", vote.Address, vote.Block) // discard the height reported by the vote
	txhash := tcrypto.Keccak256Hash([]byte(voteStr))
	txhashStr := hex.EncodeToString(txhash[:])
	return txhashStr
}
