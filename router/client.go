package router

import (
	"bytes"
	"context"
	"net/http"
	"time"
	"dist-cache/node"
)

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
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, err
	}


	req.Header.Set(
		"Content-Type",
		"application/json",
	)


	client := &http.Client{
		Timeout: 2 * time.Second,
	}


	return client.Do(req)
}
