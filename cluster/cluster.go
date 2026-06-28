package cluster


import (
	"dist-cache/node"
	"hash/fnv"
)


type Cluster struct {

	nodes []*node.Node

}

func NewCluster() *Cluster {

	return &Cluster{
		nodes: make([]*node.Node, 0),
	}
}


func (c *Cluster) AddNode(
	n *node.Node,
){

	c.nodes = append(
		c.nodes,
		n,
	)

}

func (c *Cluster) GetNode(
	key string,
) *node.Node {

	if len(c.nodes) == 0 {
		return nil
	}

	h := fnv.New32a()
	h.Write([]byte(key))
	hashValue := h.Sum32()

	index := int(hashValue) % len(c.nodes)

	return c.nodes[index]
}