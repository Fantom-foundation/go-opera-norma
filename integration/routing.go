package integration

import (
	"fmt"
	"github.com/Fantom-foundation/go-opera/integration/fakemultidb"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/cachedproducer"
	"github.com/Fantom-foundation/lachesis-base/kvdb/multidb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/skipkeys"

	"github.com/Fantom-foundation/go-opera/utils/dbutil/threads"
)

type RoutingConfig struct {
	Table map[string]multidb.Route
}

func (a RoutingConfig) Equal(b RoutingConfig) bool {
	if len(a.Table) != len(b.Table) {
		return false
	}
	for k, v := range a.Table {
		if b.Table[k] != v {
			return false
		}
	}
	return true
}

func MakeMultiProducer(rawProducers map[multidb.TypeName]kvdb.IterableDBProducer, scopedProducers map[multidb.TypeName]kvdb.FullDBProducer, cfg RoutingConfig) (kvdb.FullDBProducer, error) {
	cachedProducers := make(map[multidb.TypeName]kvdb.FullDBProducer)
	var flushID []byte
	var err error
	for typ, producer := range scopedProducers {
		flushID, err = producer.Initialize(rawProducers[typ].Names(), flushID)
		if err != nil {
			return nil, fmt.Errorf("failed to open existing databases: %v. Try to use 'db heal' to recover", err)
		}
		cachedProducers[typ] = cachedproducer.WrapAll(producer)
	}

	multi, err := fakemultidb.NewProducer(cachedProducers["pebble-fsh"])
	if err != nil {
		return nil, fmt.Errorf("failed to construct multidb: %v", err)
	}
	p := skipkeys.WrapAllProducer(multi, MetadataPrefix)
	return threads.CountedFullDBProducer(p), err
}

func MakeDirectMultiProducer(rawProducers map[multidb.TypeName]kvdb.IterableDBProducer, cfg RoutingConfig) (kvdb.FullDBProducer, error) {
	dproducers := map[multidb.TypeName]kvdb.FullDBProducer{}
	for typ, producer := range rawProducers {
		dproducers[typ] = &DummyScopedProducer{producer}
	}
	return MakeMultiProducer(rawProducers, dproducers, cfg)
}
