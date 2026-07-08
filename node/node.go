package node

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"dist-cache/cache"
)


type Node struct {
	ID string
	Address string
	Cache *cache.Cache
	healthy bool
}


type SetRequest struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	TTL     int    `json:"ttl"`
	Version int64  `json:"version"`
}


func NewNode(
	ctx context.Context,
	id string,
	address string,
) *Node {

	c := cache.NewCache()

	c.StartCleanup(
		ctx,
		1*time.Minute,
	)


	return &Node{
		ID:id,
		Address:address,
		Cache:c,
		healthy:true,
	}
}



func (n *Node) handleCache(
	w http.ResponseWriter,
	r *http.Request,
){

	switch r.Method {


	case http.MethodGet:

		n.get(
			w,
			r,
		)


	case http.MethodPost:

		n.write(
			w,
			r,
		)

	case http.MethodDelete:
    	n.write(w, r)

	default:

		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
	}

}


func (n *Node) get(
	w http.ResponseWriter,
	r *http.Request,
){

	key := r.URL.Query().Get("key")


	value, ok := n.Cache.Get(key)

	fmt.Println(
		"get request for key:",
		key,
		"value:",
		value.Value,
		"version:",
		value.Version,
		"found:",
		ok,
	)

	if !ok {

		http.NotFound(
			w,
			r,
		)

		return
	}


	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"value":   value.Value,
			"version": value.Version,
			"deleted": value.Deleted,
		},
	)
}



type WriteRequest struct {
    Key     string `json:"key"`
    Value   string `json:"value,omitempty"`
    TTL     int    `json:"ttl,omitempty"`
    Version int64  `json:"version"`
    Deleted bool   `json:"deleted"`
}

func (n *Node) write(
	w http.ResponseWriter,
	r *http.Request,
) {

	var req WriteRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	if req.Deleted {

		n.Cache.Delete(
			req.Key,
			req.Version,
		)

	} else {

		n.Cache.Set(
			req.Key,
			req.Value,
			time.Duration(req.TTL)*time.Second,
			req.Version,
		)
	}

	json.NewEncoder(w).Encode(
		map[string]string{
			"status": "stored",
		},
	)
}

// func (n *Node) delete(
// 	w http.ResponseWriter,
// 	r *http.Request,
// ){

// 	key := r.URL.Query().Get("key")


// 	n.Cache.Delete(key)


// 	w.WriteHeader(
// 		http.StatusNoContent,
// 	)
// }


func (n *Node) health(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(
		map[string]string{
			"status": "ok",
			"node":   n.ID,
		},
	)
}

func (n *Node) Start() {

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/cache",
		n.handleCache,
	)

	mux.HandleFunc(
		"/health",
		n.health,
	)

	go func() {


		fmt.Println("node running on :", n.Address)

		err := http.ListenAndServe(
			n.Address,
			mux,
		)

		if err != nil {
			panic(err)
		}

	}()

}

func (n *Node) SetHealthy(v bool) {
    n.healthy = v
}
func (n *Node) IsHealthy() bool {
	return n.healthy
}