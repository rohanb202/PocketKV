# PocketKV

A distributed in-memory key-value store written in Go featuring consistent hashing, quorum-based replication, read repair, TTL expiration, and Dockerized multi-node deployment.

> Built from scratch to explore distributed systems fundamentals including data partitioning, replication, consistency, concurrency, and fault tolerance.

---

## Features

* Consistent hashing with virtual nodes for data partitioning
* Configurable replication factor
* Quorum-based Read, Write, and Delete operations
* Versioned writes using timestamps
* Tombstone-based deletes
* Read repair for eventual consistency
* TTL-based key expiration
* Background cleanup of expired entries
* Concurrent cache operations using goroutines and synchronization primitives
* HTTP-based communication between router and cache nodes
* Health monitoring for replica selection
* Dockerized multi-node deployment with Docker Compose

---

## Architecture

```
                  Client
                     │
                     ▼
              +---------------+
              |    Router     |
              +---------------+
                │     │     │
      ┌─────────┘     │     └─────────┐
      ▼               ▼               ▼
+------------+  +------------+  +------------+
|   Node 1   |  |   Node 2   |  |   Node 3   |
| In-Memory  |  | In-Memory  |  | In-Memory  |
|   Cache    |  |   Cache    |  |   Cache    |
+------------+  +------------+  +------------+

      Consistent Hash Ring
           Replication
        Read / Write Quorums
```

The router is stateless. It determines the replica set for each key using consistent hashing and coordinates quorum operations across cache nodes.

---

# Design

## Data Partitioning

PocketKV distributes keys using **consistent hashing**.

Instead of assigning keys by modulo (`hash(key) % N`), keys are mapped onto a hash ring. Each physical node owns multiple virtual nodes to achieve a more balanced distribution.

Benefits:

* Minimal key movement when nodes join or leave
* Better load balancing
* Horizontal scalability

---

## Replication

Each key is replicated across multiple nodes.

For every request:

1. Router hashes the key.
2. Finds the primary replica.
3. Chooses the next replicas clockwise on the ring.
4. Sends requests concurrently to all replicas.

Example:

```
Replication Factor = 3

Key
 │
 ▼
Node1
 │
 ▼
Node2
 │
 ▼
Node3
```

---

## Quorum-Based Consistency

PocketKV implements quorum reads and writes.

Typical configuration:

```
Replication Factor (N) = 3

Write Quorum (W) = 2

Read Quorum (R) = 2
```

Since

```
R + W > N
```

at least one replica participating in every read contains the latest committed version.

---

## Versioning

Every write receives a monotonically increasing timestamp-based version.

```
Version = time.Now().UnixNano()
```

Replicas compare versions to determine the newest value.

Older writes are ignored.

---

## Deletes using Tombstones

Deletes are implemented as tombstones instead of immediately removing data.

Example:

```
{
    "version": 105,
    "deleted": true
}
```

This prevents deleted values from reappearing when stale replicas respond later.

---

## Read Repair

If a read observes stale replicas,

```
Replica A  Version 10
Replica B  Version 12
Replica C  Version 12
```

the router returns Version 12 to the client and asynchronously repairs Replica A in the background.

This gradually restores replica consistency without blocking reads.

---

## TTL Expiration

Each cache entry stores an expiration timestamp.

Expired entries are:

* removed lazily during reads
* periodically cleaned using a background worker

The expiration queue is maintained using a min-heap for efficient cleanup.

---

## Concurrency

Each node supports concurrent operations safely using:

* goroutines
* mutexes
* contexts
* background cleanup workers

The router performs replica requests concurrently and returns as soon as quorum is achieved.

---

## Benchmark Environment

| Component | Value |
|----------|-------|
| CPU | AMD Ryzen 9 4900H with Radeon Graphics |
| Available CPUs (WSL2) | 8 |
| Memory | 12 GB |
| OS | Ubuntu 24.04.4 LTS (WSL2) |
| Go | 1.22.2 |
| Docker | 29.0.1 |
| Deployment | Docker Compose (1 Router + 3 Cache Nodes) |
| Benchmark Tool | hey |

---

## Performance

Benchmarks were executed against a Docker Compose deployment consisting of **1 router and 3 cache nodes**.

### GET

| Concurrency | Throughput | Average Latency |
|------------:|-----------:|----------------:|
| 100 | **3,041 req/s** | **31 ms** |
| 200 | **3,330 req/s** | **58 ms** |
| 500 | **3,470 req/s** | **138 ms** |

### POST

| Concurrency | Throughput | Average Latency |
|------------:|-----------:|----------------:|
| 50 | **2,654 req/s** | **18 ms** |

> **Note:** Benchmark results were collected on local hardware and are intended for comparison purposes. Actual performance will vary depending on CPU, memory, operating system, Docker configuration, and workload characteristics.

---

# Project Structure

```
PocketKV/

cmd/
├── node/
└── router/

cache/
cluster/
node/
router/

Dockerfile
docker-compose.yml
```

---

# Running Locally

Clone the repository

```bash
git clone https://github.com/rohanb202/PocketKV.git

cd PocketKV
```

Build

```bash
go build -o node ./cmd/node

go build -o router ./cmd/router
```

Run

```bash
./node
```

and

```bash
./router
```

---

# Running with Docker

Build the image

```bash
docker build -t pocketkv .
```

Start the cluster

```bash
docker compose up -d
```

Verify

```bash
docker compose ps
```

---

# API

## PUT / POST Value

```http
POST /cache
```

```json
{
    "key":"user1",
    "value":"rohan",
    "ttl":600
}
```

---

## Read Value

```http
GET /cache?key=user1
```

Response

```json
{
    "value":"rohan",
    "version":17524234123123
}
```

---

## Delete Value

```http
DELETE /cache?key=user1
```

Returns

```
204 No Content
```

---

# Technologies

* Go
* HTTP
* Docker
* Docker Compose
* Consistent Hashing
* Goroutines
* Mutexes
* Context
* Min Heap

---

# Learning Outcomes

This project was built to gain a practical understanding of distributed systems concepts, including:

* Consistent hashing
* Replication strategies
* Quorum-based consistency
* Eventual consistency
* Tombstones
* Read repair
* Concurrent programming in Go
* Docker-based distributed deployments

---

## License

MIT
