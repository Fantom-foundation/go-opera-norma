package evmstore

import (
	"fmt"
	carmen "github.com/Fantom-foundation/Carmen/go/state"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/topicsdb"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/core/types"
	"os"
)

const nominalSize uint = 1

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	cfg StoreConfig

	mainDB kvdb.Store
	table struct {
		// API-only tables
		Receipts    kvdb.Store `table:"r"`
		TxPositions kvdb.Store `table:"x"`
		Txs         kvdb.Store `table:"X"`
	}

	EvmLogs  topicsdb.Index

	cache struct {
		TxPositions *wlru.Cache `cache:"-"` // store by pointer
		Receipts    *wlru.Cache `cache:"-"` // store by value
		EvmBlocks   *wlru.Cache `cache:"-"` // store by pointer
	}

	rlp rlpstore.Helper

	logger.Instance

	parameters carmen.Parameters
	carmenState carmen.State
	liveStateDb carmen.StateDB
}

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store, cfg StoreConfig) *Store {
	s := &Store{
		cfg:      cfg,
		mainDB:   mainDB,
		Instance: logger.New("evm-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
		parameters: cfg.StateDb,
	}

	table.MigrateTables(&s.table, s.mainDB)

	if cfg.DisableLogsIndexing {
		s.EvmLogs = topicsdb.NewDummy()
	} else {
		s.EvmLogs = topicsdb.NewWithThreadPool(mainDB)
	}
	s.initCache()

	return s
}

// Open the StateDB database (after the genesis import)
func (s *Store) Open() error {
	err := os.MkdirAll(s.parameters.Directory, 0700)
	if err != nil {
		return fmt.Errorf("failed to create carmen dir \"%s\"; %v", s.parameters.Directory, err)
	}
	s.carmenState, err = carmen.NewState(s.parameters)
	if err != nil {
		return fmt.Errorf("failed to create carmen state; %s", err)
	}
	s.liveStateDb = carmen.CreateStateDBUsing(s.carmenState)
	return nil
}

// Close closes underlying database.
func (s *Store) Close() error {
	// set all table/cache fields to nil
	table.MigrateTables(&s.table, nil)
	table.MigrateCaches(&s.cache, func() interface{} {
		return nil
	})
	s.EvmLogs.Close()

	if s.liveStateDb != nil {
		err := s.liveStateDb.Close()
		if err != nil {
			return fmt.Errorf("failed to close Carmen State: %w", err)
		}
		s.carmenState = nil
		s.liveStateDb = nil
	}
	return nil
}

func (s *Store) initCache() {
	s.cache.Receipts = s.makeCache(s.cfg.Cache.ReceiptsSize, s.cfg.Cache.ReceiptsBlocks)
	s.cache.TxPositions = s.makeCache(nominalSize*uint(s.cfg.Cache.TxPositions), s.cfg.Cache.TxPositions)
	s.cache.EvmBlocks = s.makeCache(s.cfg.Cache.EvmBlocksSize, s.cfg.Cache.EvmBlocksNum)
}

// IndexLogs indexes EVM logs
func (s *Store) IndexLogs(recs ...*types.Log) {
	err := s.EvmLogs.Push(recs...)
	if err != nil {
		s.Log.Crit("DB logs index error", "err", err)
	}
}

/*
 * Utils:
 */

func (s *Store) makeCache(weight uint, size int) *wlru.Cache {
	cache, err := wlru.New(weight, size)
	if err != nil {
		s.Log.Crit("Failed to create LRU cache", "err", err)
		return nil
	}
	return cache
}
