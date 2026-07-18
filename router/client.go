package router

import (
	"bytes"
	"context"
	"net/http"
	"time"
	"dist-cache/node"
)


var httpClient = &http.Client{
    Timeout: 2 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        MaxConnsPerHost:     100,
        IdleConnTimeout:     90 * time.Second,
    },
}

func sendToNode(
    ctx context.Context,
	n *node.Node,
	method string,
	body []byte,
	path string,
) (*http.Response, error) {

	url := "http://" + n.Address + path


	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, err
	}


	req.Header.Set(
		"Content-Type",
		"application/json",
	)


	return httpClient.Do(req)
}
