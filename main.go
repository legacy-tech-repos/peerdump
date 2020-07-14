package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	ds "github.com/ipfs/go-datastore"
	bds "github.com/ipfs/go-ds-badger"
	lds "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	// supported backends
	levelDB = "leveldb"
	badger  = "badger"
)

func main() {
	backend := flag.String("b", badger, fmt.Sprintf("backend type, currently supported: %s and %s", levelDB, badger))
	rootPath := flag.String("p", ".", "path to datastore directory")
	unsafeMode := flag.Bool("u", false, "unsafe datastore access")
	flag.Parse()

	if *unsafeMode {
		fmt.Println("[WARN] unsafe datastore access")
	}

	dStore, err := initDatastore(*backend, *rootPath, !*unsafeMode)
	if err != nil {
		fail(fmt.Errorf("init datastore: %w", err))
	}

	store := NewStore(dStore)

	info, err := dumpPeers(store)
	if err != nil {
		fail(err)
	}

	if len(info) == 0 {
		fmt.Println("No peer information found")
		return
	}

	fmt.Printf("Found %d unique peers\n\n", len(info))

	var count int
	for id, addrs := range info {
		count++
		fmt.Printf("%d. %s - found %d address(es):\n", count, id.String(), len(addrs))
		for _, addr := range addrs {
			fmt.Printf("    * [%s], TTL: %d, expires at: %d (%v)\n",
				addr.Addr.String(), addr.TTL, addr.Expiry,
				time.Unix(addr.Expiry, 0).Format("15:04:05 2006/01/02"))
		}
		fmt.Println()
	}
}

func initDatastore(backend, rootPath string, safe bool) (store ds.Datastore, err error) {
	switch backend {
	case badger:
		opts := bds.DefaultOptions
		if safe {
			opts.ReadOnly = true
		}
		store, err = bds.NewDatastore(rootPath, &opts)

	case levelDB:
		opts := lds.Options{ErrorIfMissing: true}
		if safe {
			opts.ReadOnly = true
		}
		store, err = lds.NewDatastore(rootPath, &opts)

	default:
		err = fmt.Errorf("unsupported datastore backend: %s", backend)
	}

	return
}

func dumpPeers(store *Store) (map[peer.ID][]PeerAddr, error) {
	peers, err := store.Peers()
	if err != nil {
		return nil, fmt.Errorf("loading peers from datastore: %w", err)
	}

	result := make(map[peer.ID][]PeerAddr, len(peers))
	for _, id := range peers {
		info, err := store.Addrs(id)
		if err != nil {
			return nil, fmt.Errorf("loading addresses for %s: %w", id, err)
		}
		result[id] = info
	}

	return result, nil
}

func fail(err error) {
	fmt.Printf("[ERR] %s\n", err.Error())
	os.Exit(1)
}
