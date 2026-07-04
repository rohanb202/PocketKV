package cluster

import (
	"hash/fnv"
	"sort"
    "strconv"
	"dist-cache/node"
	"log/slog"
)


type RingNode struct {

	Hash uint32

	Node *node.Node

	VirtualID int

}

type HashRing struct {

	nodes []RingNode

	virtualNodes int
}

func NewHashRing() *HashRing {

	return &HashRing{
		nodes: make([]RingNode,0),
		virtualNodes:100,
	}
}

func hashKey(
	key string,
) uint32 {


	h := fnv.New32a()


	h.Write(
		[]byte(key),
	)


	return h.Sum32()
}

func (r *HashRing) AddNode(
	n *node.Node,
){

	for i:=0;i<r.virtualNodes;i++ {


		key :=
			n.ID +
			"-" +
			strconv.Itoa(i)


		hash :=
			hashKey(key)


		r.nodes = append(
			r.nodes,
			RingNode{
				Hash:hash,
				Node:n,
				VirtualID:i,
			},
		)
	}


	sort.Slice(
		r.nodes,
		func(i,j int) bool {
			return r.nodes[i].Hash <
			       r.nodes[j].Hash
		},
	)
}

func (r *HashRing) GetNodes(
	key string,
	count int,
) []*node.Node {


	hash := hashKey(key)


	start := -1


	for i,n := range r.nodes {

		if hash <= n.Hash {
			start=i
			break
		}
	}

	if start == -1 {
		start = 0
	}


	result := make([]*node.Node,0)

	seen := map[string]bool{}

	slog.Info(
		"hash for key",
		slog.String("key", key),
		slog.Int("hash", int(hash)),
	)

	for i:=0; len(result)<count; i++ {


		index :=
			(start+i)%len(r.nodes)


		n :=
			r.nodes[index].Node


		if !seen[n.ID] {

			result = append(
				result,
				n,
			)

			seen[n.ID]=true
		}
	}

	slog.Info(
		"nodes responsible for key",
		slog.String("key", key),
		slog.Any("nodes", result),
	)

	return result
}


func (r *HashRing) GetRingOrder(key string) []*node.Node {

	hash := hashKey(key)


	start := -1


	for i,n := range r.nodes {

		if hash <= n.Hash {
			start=i
			break
		}
	}

	if start == -1 {
		start = 0
	}


	result := make([]*node.Node,0)

	seen := map[string]bool{}

	slog.Info(
		"hash for key",
		slog.String("key", key),
		slog.Int("hash", int(hash)),
	)


	for i:=0; i<len(r.nodes); i++ {

		index :=
			(start+i)%len(r.nodes)
		
		n :=
			r.nodes[index].Node
		
		if !seen[n.ID] {

			result = append(
				result,
				n,
			)

			seen[n.ID]=true
		}
		
	}

	slog.Info(
		"nodes responsible for key",
		slog.String("key", key),
		slog.Any("nodes", result),
	)
	
	return result


}