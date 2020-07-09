package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	badger "github.com/ipfs/go-ds-badger"
	"github.com/libp2p/go-libp2p-core/peer"
)

const defaultIpfsLitePath = "ipfslite"

func main() {
	var dir string
	if input := os.Args; len(input) != 2 {
		fail(usage(input[0]))
	} else {
		dir = input[1]
	}

	rootPath := filepath.Join(dir, defaultIpfsLitePath)
	bds, err := badger.NewDatastore(rootPath, &badger.DefaultOptions)
	if err != nil {
		fail(fmt.Sprintf("[ERR] initializing badger datastore: %v\n", err))
	}

	store := NewStore(bds)

	info, err := dumpPeers(store)
	if err != nil {
		fail(err.Error())
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

func usage(cmdName string) string {
	return fmt.Sprintf("Usage: %s <root-path>\n", cmdName)
}

func fail(reason string) {
	if len(reason) > 0 {
		fmt.Println(reason)
	}

	os.Exit(1)
}
