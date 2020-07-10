# peerdump

Utility for dumping address information from the libp2p-based peerstore.

### Installation

```shell script
go get github.com/anytypeio/peerdump
```

### Usage

Run utility with a flag `-p` pointing to the data directory.

```shell script
peerdump -p <root-path>
```

Optional flag `-b` sets type of underlying storage backend. Currently, it supports two backend: *Badger* and *LevelDB*.
