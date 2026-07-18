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
		MaxIdleConnsPerHost:   200,  // bump from 100
		MaxConnsPerHost:       200,
		IdleConnTimeout:       90 * time.Second,
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     true,  // try HTTP/2 automatically
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
