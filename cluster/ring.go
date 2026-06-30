package cluster

import (
	"hash/fnv"
	"sort"
"strconv"
	"dist-cache/node"
	"fmt"
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


	start := 0


	for i,n := range r.nodes {

		if hash <= n.Hash {
			start=i
			break
		}
	}


	result := make([]*node.Node,0)

	seen := map[string]bool{}

	fmt.Println("hash for key",key,"is",hash)

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

	fmt.Println("nodes responsible for key",key,"are",result)
	return result
}