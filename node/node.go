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
	Key string `json:"key"`
	Value string `json:"value"`
	TTL int `json:"ttl"`
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

		n.set(
			w,
			r,
		)


	case http.MethodDelete:

		n.delete(
			w,
			r,
		)


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
		value,
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
		map[string]string{
			"value":value,
		},
	)
}



func (n *Node) set(
	w http.ResponseWriter,
	r *http.Request,
){

	var req SetRequest


	err := json.NewDecoder(
		r.Body,
	).Decode(&req)


	if err != nil {
		http.Error(
			w,
			err.Error(),
			400,
		)
		return
	}

	fmt.Println(
		"set request for key:",
		req.Key,
		"value:",
		req.Value,
		"ttl:",
		req.TTL,
	)
	n.Cache.Set(
		req.Key,
		req.Value,
		time.Duration(req.TTL)*time.Second,
	)


	json.NewEncoder(w).Encode(
		map[string]string{
			"status":"stored",
		},
	)
}


func (n *Node) delete(
	w http.ResponseWriter,
	r *http.Request,
){

	key := r.URL.Query().Get("key")


	n.Cache.Delete(key)


	w.WriteHeader(
		http.StatusNoContent,
	)
}


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