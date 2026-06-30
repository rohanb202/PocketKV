package cluster


import (
	"dist-cache/node"

)


type Cluster struct {

	ring *HashRing

}

func NewCluster() *Cluster {

	return &Cluster{

		ring:&HashRing{
			nodes:make([]RingNode,0),
			virtualNodes:100,
		},
	}
}

func (c *Cluster) AddNode(
	n *node.Node,
){

	c.ring.AddNode(n)

}

func (c *Cluster) GetNodes(
	key string,
	count int,
) []*node.Node {


	return c.ring.GetNodes(key, count)

}