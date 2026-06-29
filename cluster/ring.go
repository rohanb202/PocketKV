package cluster

import (
	"hash/fnv"
	"sort"

	"dist-cache/node"
)


type RingNode struct {

	Hash uint32

	Node *node.Node

}


type HashRing struct {

	nodes []RingNode

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
) {


	hash := hashKey(
		n.ID,
	)


	r.nodes = append(
		r.nodes,
		RingNode{
			Hash: hash,
			Node:n,
		},
	)


	sort.Slice(
		r.nodes,
		func(i,j int) bool {

			return r.nodes[i].Hash <
			       r.nodes[j].Hash
		},
	)
}

func (r *HashRing) GetNode(
	key string,
) *node.Node {


	if len(r.nodes)==0 {
		return nil
	}


	hash := hashKey(key)


	for _, n := range r.nodes {


		if hash <= n.Hash {

			return n.Node
		}
	}


	// wrap around
	return r.nodes[0].Node
}