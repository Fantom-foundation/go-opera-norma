package inter

import (
	"crypto/sha256"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Block struct {
	Time        Timestamp
	Atropos     hash.Event
	Events      hash.Events
	Txs         []common.Hash // non event txs (received via genesis or LLR)
	InternalTxs []common.Hash // DEPRECATED in favor of using only Txs fields and method internal.IsInternal
	SkippedTxs  []uint32      // indexes of skipped txs, starting from first tx of first event, ending with last tx of last event
	GasUsed     uint64
	Root        hash.Hash
	prevRandao  common.Hash
}

func (b *Block) EstimateSize() int {
	return (len(b.Events)+len(b.InternalTxs)+len(b.Txs)+1+1)*32 + len(b.SkippedTxs)*4 + 8 + 8
}

// GetPrevRandao returns the PrevRandao for current block.
func (b *Block) GetPrevRandao() common.Hash {
	if b.prevRandao == (common.Hash{}) {
		b.computePrevRandao()
	}
	return b.prevRandao
}

// computePrevRandao computes the PrevRandao from event hashes.
func (b *Block) computePrevRandao() {
	// reset the value
	b.prevRandao = common.Hash{}
	for _, event := range b.Events {
		for i := 0; i < 24; i++ {
			// first 8 bytes should be ignored as they are not pseudo-random.
			b.prevRandao[i+8] = b.prevRandao[i+8] ^ event[i+8]
		}
	}
	b.prevRandao = sha256.Sum256(b.prevRandao.Bytes())
}

func FilterSkippedTxs(txs types.Transactions, skippedTxs []uint32) types.Transactions {
	if len(skippedTxs) == 0 {
		// short circuit if nothing to skip
		return txs
	}
	skipCount := 0
	filteredTxs := make(types.Transactions, 0, len(txs))
	for i, tx := range txs {
		if skipCount < len(skippedTxs) && skippedTxs[skipCount] == uint32(i) {
			skipCount++
		} else {
			filteredTxs = append(filteredTxs, tx)
		}
	}

	return filteredTxs
}
