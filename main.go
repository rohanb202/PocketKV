package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"context"
	"dist-cache/cache"
)


var c *cache.Cache


func getValue(
	w http.ResponseWriter,
	r *http.Request,
){

	key := r.URL.Query().Get("key")


	value, ok := c.Get(key)


	if !ok {

		http.Error(
			w,
			"not found",
			404,
		)

		return
	}


	json.NewEncoder(w).Encode(
		map[string]string{
			"value":value,
		},
	)

}
func deleteValue(
	w http.ResponseWriter,
	r *http.Request,
){

	key := r.URL.Query().Get("key")


	c.Delete(key)


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



	c.Set(
		req.Key,
		req.Value,
		time.Duration(req.TTL)*time.Second,
	)


	w.WriteHeader(
		http.StatusCreated,
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

	c = cache.NewCache()

	ctx,cancel := context.WithCancel(
		context.Background(),
	)

	defer cancel()


	c.StartCleanup(
		ctx,
		time.Minute,
	)

	http.HandleFunc(
		"/cache",
		cacheHandler,
	)


	fmt.Println("server running on :8080")


	http.ListenAndServe(
		":8080",
		nil,
	)

}