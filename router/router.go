package router


import (
	"dist-cache/cluster"
)

type Router struct {
	Cluster *cluster.Cluster
}

func NewRouter(
	cluster *cluster.Cluster,
) *Router {
	return &Router{
		Cluster: cluster,
	}
}




