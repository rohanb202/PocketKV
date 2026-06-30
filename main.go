package main

import (
	"encoding/json"
	"net/http"
	"context"
	"dist-cache/node"
	"dist-cache/cluster"
	"bytes"
	"fmt"
	"io"
	"time"
)


var cl *cluster.Cluster

func sendToNode(
	n *node.Node,
	method string,
	body []byte,
	path string,
) (*http.Response, error) {

	url := "http://" + n.Address + path


	req, err := http.NewRequest(
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

func getValue(
	w http.ResponseWriter,
	r *http.Request,
) {

	key := r.URL.Query().Get("key")


	nodes := cl.GetNodes(
		key,
		2,
	)


	for _, n := range nodes {

		resp, err := sendToNode(
			n,
			http.MethodGet,
			nil,
			"/cache?key="+key,
		)


		if err != nil {
			continue
		}


		// close immediately after handling this response
		if resp.StatusCode == http.StatusOK {

			w.WriteHeader(
				http.StatusOK,
			)

			_, _ = io.Copy(
				w,
				resp.Body,
			)

			resp.Body.Close()

			return
		}


		resp.Body.Close()
	}


	http.Error(
		w,
		"key not found",
		http.StatusNotFound,
	)
}

func deleteValue(
	w http.ResponseWriter,
	r *http.Request,
) {

	key := r.URL.Query().Get("key")


	nodes := cl.GetNodes(
		key,
		2,
	)


	success := 0


	for _, n := range nodes {


		resp, err := sendToNode(
			n,
			http.MethodDelete,
			nil,
			"/cache?key="+key,
		)


		if err != nil {
			continue
		}


		resp.Body.Close()


		if resp.StatusCode == http.StatusNoContent {
			success++
		}
	}


	if success == 0 {

		http.Error(
			w,
			"delete failed",
			http.StatusServiceUnavailable,
		)

		return
	}


	w.WriteHeader(
		http.StatusNoContent,
	)
}
type SetRequest struct {

	Key string `json:"key"`

	Value string `json:"value"`

	TTL int `json:"ttl"`
}



func setValue(
	w http.ResponseWriter,
	r *http.Request,
) {

	body, err := io.ReadAll(r.Body)
	
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}


	var req SetRequest

	err = json.Unmarshal(
		body,
		&req,
	)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
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

	nodes := cl.GetNodes(
		req.Key,
		2,
	)

	// fmt.Println(
	// 	"nodes responsible for key:",
	// 	req.Key,nodes
	// )


	success := 0


	for _, n := range nodes {

		fmt.Println(
			"sending to",
			n.Address,
		)
		resp, err := sendToNode(
			n,
			http.MethodPost,
			body,
			"/cache",
		)


		if err != nil {
			continue
		}


		resp.Body.Close()


		if resp.StatusCode >= 200 &&
		   resp.StatusCode < 300 {

			success++
		}
	}


	if success == 0 {

		http.Error(
			w,
			"all replicas failed",
			http.StatusServiceUnavailable,
		)

		return
	}


	w.WriteHeader(
		http.StatusCreated,
	)


	json.NewEncoder(w).Encode(
		map[string]int{
			"replicas_written":success,
		},
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


	cl.AddNode(n1)
	cl.AddNode(n2)

	n1.Start()
	n2.Start()


	

   fmt.Println("router running on :8080")

	http.HandleFunc(
		"/cache",
		cacheHandler,
	)


	http.ListenAndServe(
		":8080",
		nil,
	)
}