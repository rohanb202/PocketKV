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
)


var cl *cluster.Cluster

func sendToNode(
	node *node.Node,
	method string,
	body []byte,
	path string,
) (*http.Response,error){

	url := "http://" + node.Address + path


	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(body),
	)


	if err != nil {
		return nil,err
	}


	req.Header.Set(
		"Content-Type",
		"application/json",
	)


	client := &http.Client{}


	return client.Do(req)
}


func getValue(
	w http.ResponseWriter,
	r *http.Request,
){

	key := r.URL.Query().Get("key")


	n := cl.GetNode(key)


	resp, err := sendToNode(
		n,
		http.MethodGet,
		nil,
		"/cache?key="+key,
	)


	if err != nil {
		http.Error(
			w,
			err.Error(),
			500,
		)
		return
	}


	defer resp.Body.Close()


	w.WriteHeader(resp.StatusCode)


	_,_ = io.Copy(
		w,
		resp.Body,
	)
}

func deleteValue(
	w http.ResponseWriter,
	r *http.Request,
){

	key := r.URL.Query().Get("key")


	n := cl.GetNode(key)


	resp, err := sendToNode(
		n,
		http.MethodDelete,
		nil,
		"/cache?key="+key,
	)


	if err != nil {
		http.Error(
			w,
			err.Error(),
			500,
		)
		return
	}


	defer resp.Body.Close()


	w.WriteHeader(resp.StatusCode)
}

type SetRequest struct {

	Key string `json:"key"`

	Value string `json:"value"`

	TTL int `json:"ttl"`
}



func setValue(
	w http.ResponseWriter,
	r *http.Request,
){

	body, err := io.ReadAll(r.Body)


	if err != nil {
		http.Error(w,err.Error(),400)
		return
	}


	var req SetRequest


	json.Unmarshal(
		body,
		&req,
	)


	n := cl.GetNode(req.Key)



	resp, err := sendToNode(
		n,
		http.MethodPost,
		body,
		"/cache",
	)


	if err != nil {
		http.Error(
			w,
			err.Error(),
			500,
		)
		return
	}


	defer resp.Body.Close()


	w.WriteHeader(resp.StatusCode)
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


	n1.Start()
	n2.Start()


	cl.AddNode(n1)
	cl.AddNode(n2)

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