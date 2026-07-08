package cluster


import (
	"dist-cache/node"
	"time"
	"context"

)


type Cluster struct {

	ring *HashRing
	nodes []*node.Node
	checker *HealthChecker
	replicationFactor int
	writeQuorum int
	readQuorum int

}

func NewCluster() *Cluster {

	return &Cluster{

		ring:&HashRing{
			nodes:make([]RingNode,0),
			virtualNodes:100,
		},
		nodes: make([]*node.Node, 0),
		checker: NewHealthChecker(2 * time.Second),
		replicationFactor: 3,
		writeQuorum: 2,
		readQuorum: 2,
	}
}


func (c *Cluster) ReplicationFactor() int {
    return c.replicationFactor
}

func (c *Cluster) WriteQuorum() int {
    return c.writeQuorum
}

func (c *Cluster) ReadQuorum() int {
	return c.readQuorum
}



func (c *Cluster) AddNode(
	n *node.Node,
){
	c.nodes = append(c.nodes, n)
	c.ring.AddNode(n)

}

func (c *Cluster) GetNodes(
	key string,
	count int,
) []*node.Node {


	return c.ring.GetNodes(key, count)

}

func (c *Cluster) GetHealthyNodes(key string,count int) []*node.Node {

	healthyNodes := make([]*node.Node, 0)
	
	nodes := c.ring.GetRingOrder(key)

	for _, n := range nodes {
		if n.IsHealthy() {
			healthyNodes = append(healthyNodes, n)
		}
		if len(healthyNodes) == count {
			break
		}
	}

	return healthyNodes

}

func (c *Cluster) Start(
    ctx context.Context,
) {
    go c.checker.Start(
        ctx,
        2*time.Second,
        c.nodes,
    )
}