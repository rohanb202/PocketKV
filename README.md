# PocketKV

A distributed in-memory key-value cache built in **Go** to explore the fundamentals of distributed systems. PocketKV implements consistent hashing, replication, quorum-based consistency, tombstones, read repair, and concurrent request handling using Go's concurrency primitives.

## Features

* Distributed in-memory key-value store
* Consistent hashing with virtual nodes
* Configurable replication factor
* Read, write, and delete quorums
* Version-based conflict resolution
* Tombstones for safe deletes
* Asynchronous read repair
* TTL-based key expiration
* Background cleanup workers
* Health checking of cache nodes
* Parallel replication using goroutines and channels
* Context-aware request cancellation
* HTTP-based inter-node communication

---

## Architecture

```
                   +----------------------+
                   |        Router        |
                   |----------------------|
                   | Consistent Hash Ring |
                   | Quorum Coordinator   |
                   +----------+-----------+
                              |
              -------------------------------------
             |                  |                  |
     +-------v------+   +-------v------+   +-------v------+
     |    Node 1    |   |    Node 2    |   |    Node 3    |
     |--------------|   |--------------|   |--------------|
     | In-Memory KV |   | In-Memory KV |   | In-Memory KV |
     | TTL Cleanup  |   | TTL Cleanup  |   | TTL Cleanup  |
     +--------------+   +--------------+   +--------------+
```

The router is responsible for:

* Consistent hash lookup
* Selecting replica nodes
* Coordinating read/write/delete quorums
* Performing read repair
* Monitoring node health

Each cache node:

* Stores data locally
* Maintains key versions
* Handles TTL expiration
* Processes HTTP requests independently

---

## Consistency Model

PocketKV uses **eventual consistency** with quorum replication.

### Write

* Router assigns a version to every write.
* Data is replicated to multiple nodes in parallel.
* Request succeeds after the configured write quorum is reached.

### Read

* Router queries replicas concurrently.
* Waits for the configured read quorum.
* Returns the latest version.
* Repairs stale replicas asynchronously.

### Delete

Deletes are implemented using **tombstones** instead of immediate removal to prevent deleted values from reappearing due to stale replicas.

---

## Technologies

* Go
* HTTP
* Goroutines
* Channels
* Mutexes
* Context
* Heap (TTL expiration)
* Consistent Hashing

---

## Running

Start three cache nodes:

```bash
go run ./cmd/node
```

Configure each node using environment variables:

```text
NODE_ID=node1
NODE_ADDRESS=:8081
```

Start the router:

```bash
go run ./cmd/router
```

Router configuration:

```text
ROUTER_PORT=8080
NODES=localhost:8081,localhost:8082,localhost:8083
```

---

## Example

Store a value

```bash
curl -X POST localhost:8080/cache \
-H "Content-Type: application/json" \
-d '{"key":"name","value":"PocketKV","ttl":60}'
```

Read a value

```bash
curl localhost:8080/cache?key=name
```

Delete a value

```bash
curl -X DELETE localhost:8080/cache?key=name
```

---

## Project Structure

```
PocketKV/
├── cmd/
│   ├── node/
│   └── router/
├── cache/
├── cluster/
├── node/
├── router/
├── hashing/
└── README.md
```

---

## Future Improvements

* Docker Compose deployment
* Load testing and benchmarking
* Hinted handoff
* Merkle-tree based anti-entropy
* Vector clocks for multi-writer conflict resolution
* Gossip-based membership
* Persistent storage (WAL + snapshots)
* Metrics with Prometheus and Grafana
* gRPC communication between nodes

---

## Inspiration

PocketKV is inspired by the design principles behind distributed databases and caches such as Amazon Dynamo, Apache Cassandra, and Redis Cluster. It is intended as a learning project focused on distributed systems concepts rather than a production-ready datastore.
