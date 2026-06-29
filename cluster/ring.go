package cluster

import (
	"hash/fnv"
	"sort"
"strconv"
	"dist-cache/node"
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