package router

import (
	"dist-cache/node"
	"encoding/json"
	"context"
	"net/http"
	"time"
	"log/slog"
	"fmt"

)
type ReadResult struct {
	Node    *node.Node
    Found   bool
    Value   string
    Version int64
    Err     error
	Deleted bool
}

func (rt *Router) getValue(
	w http.ResponseWriter,
	r *http.Request,
) {  

	key := r.URL.Query().Get("key")

	nodes := rt.Cluster.GetHealthyNodes(
		key,
		rt.Cluster.ReplicationFactor(),
	)

	if len(nodes) == 0 {
		http.Error(
			w,
			"no healthy replicas",
			http.StatusServiceUnavailable,
		)
		return
	}

	readQuorum := rt.Cluster.ReadQuorum()

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
				Node:    node,
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



type WriteRequest struct {
    Key     string `json:"key"`
    Value   string `json:"value,omitempty"`
    TTL     int    `json:"ttl,omitempty"`
    Version int64  `json:"version"`
    Deleted bool   `json:"deleted"`
}

func (rt *Router) replicateWrite(
	ctx context.Context,
	req WriteRequest,
) (int, error) {

	nodes := rt.Cluster.GetHealthyNodes(
		req.Key,
		rt.Cluster.ReplicationFactor(),
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

	writeQuorum := rt.Cluster.WriteQuorum()

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


func (rt *Router) setValue(
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

	success, err := rt.replicateWrite(r.Context(), req)

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


func (rt *Router) deleteValue(
	w http.ResponseWriter,
	r *http.Request,
) {

	req := WriteRequest{
		Key:     r.URL.Query().Get("key"),
		Version: time.Now().UnixNano(),
		Deleted: true,
	}

	success, err := rt.replicateWrite(r.Context(), req)

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


func (rt *Router) CacheHandler(
	w http.ResponseWriter,
	r *http.Request,
){

	switch r.Method {


		case http.MethodPost:

			rt.setValue(w,r)


		case http.MethodGet:

			rt.getValue(w,r)


		case http.MethodDelete:

			rt.deleteValue(w,r)


		default:

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
	}

}
