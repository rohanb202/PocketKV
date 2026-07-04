package cluster

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"dist-cache/node"
)

type HealthChecker struct {
    timeout time.Duration
    client  *http.Client
}

func NewHealthChecker(
    timeout time.Duration,
) *HealthChecker {
	return &HealthChecker{
		timeout: timeout,
		client:  &http.Client{},
	}
}

func (hc *HealthChecker) CheckHealth(
    nodes []*node.Node,
) {
    for _, n := range nodes {
        hc.checkNodeHealth(n)
    }
}

func (hc *HealthChecker) checkNodeHealth(n *node.Node) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://"+n.Address+"/health", nil)
	if err != nil {
		fmt.Printf("Error creating request for node %s: %v\n", n.Address, err)
		return
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		fmt.Printf("Node %s is unhealthy: %v\n", n.Address, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		n.SetHealthy(true)
		fmt.Printf("Node %s is healthy\n", n.Address)
	} else {
		n.SetHealthy(false)
		fmt.Printf("Node %s is unhealthy: status code %d\n", n.Address, resp.StatusCode)
	}
}


func (hc *HealthChecker) Start(
    ctx context.Context,
    interval time.Duration,
    nodes []*node.Node,
) {

    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {

        select {

        case <-ticker.C:

            hc.CheckHealth(nodes)

        case <-ctx.Done():

            return
        }
    }
}