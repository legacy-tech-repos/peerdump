package main

import (
	"errors"
	"fmt"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/libp2p/go-libp2p-core/peer"
	pstore_pb "github.com/libp2p/go-libp2p-peerstore/pb"
	"github.com/multiformats/go-base32"
	b32 "github.com/multiformats/go-base32"
	"github.com/multiformats/go-multiaddr"
)

// Must match current scheme defined in github.com/libp2p/go-libp2p-peerstore/pstoreds/addr_book.go
// Peer addresses are stored db key pattern: /peers/addrs/<b32 peer id no padding>
var addrBookBase = ds.NewKey("/peers/addrs")

type PeerAddr struct {
	Addr   multiaddr.Multiaddr
	Expiry int64
	TTL    int64
}

type Store struct {
	ds ds.Datastore
}

func NewStore(ds ds.Datastore) *Store {
	return &Store{ds: ds}
}

// Return unique set of known peers.
func (s Store) Peers() (peer.IDSlice, error) {
	return s.uniquePeerIds(func(result query.Result) string { return ds.RawKey(result.Key).Name() })
}

// Return information about known addresses of a given peer.
func (s Store) Addrs(id peer.ID) ([]PeerAddr, error) {
	rec, err := s.loadRecord(id)
	if err != nil {
		return nil, err
	} else if rec == nil {
		return nil, nil
	}

	addrs := make([]PeerAddr, len(rec.Addrs))
	for i, entry := range rec.Addrs {
		addrs[i] = PeerAddr{
			Addr:   entry.Addr.Multiaddr,
			Expiry: entry.Expiry,
			TTL:    entry.Ttl,
		}
	}

	return addrs, nil
}

func (s Store) uniquePeerIds(extractor func(result query.Result) string) (peer.IDSlice, error) {
	var (
		q       = query.Query{Prefix: addrBookBase.String(), KeysOnly: true}
		results query.Results
		err     error
	)

	if results, err = s.ds.Query(q); err != nil {
		return nil, err
	}

	defer results.Close()

	idset := make(map[string]struct{})
	for result := range results.Next() {
		k := extractor(result)
		idset[k] = struct{}{}
	}

	ids := make(peer.IDSlice, 0, len(idset))
	for id := range idset {
		pid, _ := base32.RawStdEncoding.DecodeString(id)
		id, _ := peer.IDFromBytes(pid)
		ids = append(ids, id)
	}

	return ids, nil
}

func (s Store) loadRecord(id peer.ID) (*pstore_pb.AddrBookRecord, error) {
	var (
		pr  pstore_pb.AddrBookRecord
		key = addrBookBase.ChildString(b32.RawStdEncoding.EncodeToString([]byte(id)))
	)

	if data, err := s.ds.Get(key); err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			pr.Id = &pstore_pb.ProtoPeerID{ID: id}
			return &pr, nil
		}
		return nil, fmt.Errorf("retrieving address book record: %w", err)
	} else if err = pr.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("decoding address book record: %w", err)
	}

	return &pr, nil
}
