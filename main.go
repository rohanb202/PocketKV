package main

import (
	"encoding/json"
	"net/http"
	"context"
	"dist-cache/node"
	"dist-cache/cluster"
	"bytes"
	"time"
	"log/slog"
	"fmt"
)


var cl *cluster.Cluster


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


type ReadResult struct {
	Node    *node.Node
    Found   bool
    Value   string
    Version int64
    Err     error
	Deleted bool
}

func getValue(
	w http.ResponseWriter,
	r *http.Request,
) {

	key := r.URL.Query().Get("key")

	nodes := cl.GetHealthyNodes(
		key,
		cl.ReplicationFactor(),
	)

	if len(nodes) == 0 {
		http.Error(
			w,
			"no healthy replicas",
			http.StatusServiceUnavailable,
		)
		return
	}

	readQuorum := cl.ReadQuorum()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	resultChan := make(chan ReadResult, len(nodes))

	for _, n := range nodes {

		go func(node *node.Node) {

			resp, err := sendToNode(
				ctx,
				node,
				http.MethodGet,
				nil,
				"/cache?key="+key,
			)

			if err != nil {

				select {
				case resultChan <- ReadResult{Err: err}:
				case <-ctx.Done():
				}

				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {

				select {
				case resultChan <- ReadResult{Found: false}:
				case <-ctx.Done():
				}

				return
			}

			var data struct {
				Value   string `json:"value"`
				Version int64  `json:"version"`
				Deleted bool   `json:"deleted"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {

				select {
				case resultChan <- ReadResult{Err: err}:
				case <-ctx.Done():
				}

				return
			}

			select {

			case resultChan <- ReadResult{
				Found:   true,
				Value:   data.Value,
				Version: data.Version,
				Deleted: data.Deleted,
			}:

			case <-ctx.Done():
			}

		}(n)
	}

	success := make([]ReadResult, 0, readQuorum)
	failure := 0

	for i := 0; i < len(nodes); i++ {

		result := <-resultChan

		if result.Err != nil || !result.Found {
			failure=failure+1
			if failure > len(nodes)-readQuorum {
				http.Error(
					w,
					"read quorum not reached",
					http.StatusServiceUnavailable,
				)
				return
			}
		} else {
			success = append(success, result)

			if len(success) >= readQuorum {

				// Stop all remaining requests
				cancel()

				break
			}
		}
		
	}

	if len(success) < readQuorum {

		http.Error(
			w,
			"read quorum not reached",
			http.StatusServiceUnavailable,
		)

		return
	}

	latest := success[0]

	for _, r := range success {

		if r.Version > latest.Version {
			latest = r
		}
	}


	for _, replica := range success {

		if replica.Version < latest.Version {

			go repairReplica(context.Background(), replica.Node, WriteRequest{
				Key:     key,
				Value:   latest.Value,
				Version: latest.Version,
				Deleted: latest.Deleted,
			})
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if latest.Deleted {
		http.Error(
			w,
			"key not found",
			http.StatusNotFound,
		)
		return
	}

	slog.Info(
		"get request for key:",
		key,
		"value:",
		latest.Value,
		"version:",
		latest.Version,
	)

	json.NewEncoder(w).Encode(
		map[string]any{
			"value":   latest.Value,
			"version": latest.Version,
		},
	)
}


func repairReplica(
	ctx context.Context,
	n *node.Node,
	req WriteRequest,
) {

	body, err := json.Marshal(req)
	if err != nil {
		return
	}

	resp, err := sendToNode(
		ctx,
		n,
		http.MethodPost,
		body,
		"/cache",
	)

	if err != nil {
		return
	}

	defer resp.Body.Close()
}

// func deleteValue(
// 	w http.ResponseWriter,
// 	r *http.Request,
// ) {

// 	key := r.URL.Query().Get("key")

// 	nodes := cl.GetHealthyNodes(
// 		key,
// 		cl.ReplicationFactor(),
// 	)

// 	if len(nodes) == 0 {
// 		http.Error(
// 			w,
// 			"no healthy replicas",
// 			http.StatusServiceUnavailable,
// 		)
// 		return
// 	}

// 	ctx, cancel := context.WithCancel(r.Context())
// 	defer cancel()

// 	resultChan := make(chan bool, len(nodes))

// 	for _, n := range nodes {

// 		go func(node *node.Node) {

// 			resp, err := sendToNode(
// 				ctx,
// 				node,
// 				http.MethodDelete,
// 				nil,
// 				"/cache?key="+key,
// 			)

// 			if err != nil {
// 				select {
// 				case resultChan <- false:
// 				case <-ctx.Done():
// 				}
// 				return
// 			}

// 			defer resp.Body.Close()

// 			ok := resp.StatusCode >= http.StatusOK &&
// 				resp.StatusCode < http.StatusMultipleChoices

// 			select {
// 			case resultChan <- ok:
// 			case <-ctx.Done():
// 			}

// 		}(n)
// 	}

// 	success := 0
// 	failure := 0

// 	deleteQuorum := cl.DeleteQuorum()

// 	for i := 0; i < len(nodes); i++ {

// 		ok := <-resultChan

// 		if ok {

// 			success++

// 			if success >= deleteQuorum {

// 				cancel()

// 				w.WriteHeader(http.StatusNoContent)

// 				return
// 			}

// 		} else {

// 			failure++

// 			if failure > len(nodes)-deleteQuorum {

// 				cancel()

// 				http.Error(
// 					w,
// 					"delete quorum not reached",
// 					http.StatusServiceUnavailable,
// 				)

// 				return
// 			}
// 		}
// 	}
// }



type WriteRequest struct {
    Key     string `json:"key"`
    Value   string `json:"value,omitempty"`
    TTL     int    `json:"ttl,omitempty"`
    Version int64  `json:"version"`
    Deleted bool   `json:"deleted"`
}

func replicateWrite(
	ctx context.Context,
	req WriteRequest,
) (int, error) {

	nodes := cl.GetHealthyNodes(
		req.Key,
		cl.ReplicationFactor(),
	)

	if len(nodes) == 0 {
		return 0, fmt.Errorf("no healthy replicas")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultChan := make(chan bool, len(nodes))

	for _, n := range nodes {

		go func(node *node.Node) {

			resp, err := sendToNode(
				ctx,
				node,
				http.MethodPost,
				body,
				"/cache",
			)

			if err != nil {
				select {
				case resultChan <- false:
				case <-ctx.Done():
				}
				return
			}

			defer resp.Body.Close()

			ok := resp.StatusCode >= http.StatusOK &&
				resp.StatusCode < http.StatusMultipleChoices

			select {
			case resultChan <- ok:
			case <-ctx.Done():
			}

		}(n)
	}

	success := 0
	failure := 0

	writeQuorum := cl.WriteQuorum()

	for i := 0; i < len(nodes); i++ {

		ok := <-resultChan

		if ok {

			success++

			if success >= writeQuorum {
				cancel()
				return success, nil
			}

		} else {

			failure++

			if failure > len(nodes)-writeQuorum {
				cancel()
				return success, fmt.Errorf("write quorum not reached")
			}
		}
	}

	return success, fmt.Errorf("write quorum not reached")
}


func setValue(
	w http.ResponseWriter,
	r *http.Request,
) {

	var req WriteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Version = time.Now().UnixNano()
	req.Deleted = false

	success, err := replicateWrite(r.Context(), req)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusServiceUnavailable,
		)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]any{
		"replicas_written": success,
		"quorum":           true,
	})
}


func deleteValue(
	w http.ResponseWriter,
	r *http.Request,
) {

	req := WriteRequest{
		Key:     r.URL.Query().Get("key"),
		Version: time.Now().UnixNano(),
		Deleted: true,
	}

	success, err := replicateWrite(r.Context(), req)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusServiceUnavailable,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	slog.Info(
		"delete quorum reached",
		slog.Int("replicas", success),
	)
}


func cacheHandler(
	w http.ResponseWriter,
	r *http.Request,
){

	switch r.Method {


		case http.MethodPost:

			setValue(w,r)


		case http.MethodGet:

			getValue(w,r)


		case http.MethodDelete:

			deleteValue(w,r)


		default:

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
	}

}




func main(){

	ctx,cancel := context.WithCancel(
		context.Background(),
	)

	defer cancel()


	cl = cluster.NewCluster()


	n1 := node.NewNode(
		ctx,
		"node1",
		":8081",
	)


	n2 := node.NewNode(
		ctx,
		"node2",
		":8082",
	)

	n3 := node.NewNode(
		ctx,
		"node3",
		":8084",
	)


	cl.AddNode(n1)
	cl.AddNode(n2)
	cl.AddNode(n3)

	cl.Start(ctx)

	n1.Start()
	n2.Start()
	n3.Start()

	

    slog.Info("router running on :8080" )

	http.HandleFunc(
		"/cache",
		cacheHandler,
	)


	http.ListenAndServe(
		":8080",
		nil,
	)
}