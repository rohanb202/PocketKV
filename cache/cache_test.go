package cache

import (
	"testing"
	"time"
)


func TestSetGet(t *testing.T){

	c := NewCache()


	c.Set(
		"name",
		"rohan",
		5*time.Second,
	)


	value, ok := c.Get("name")


	if !ok {
		t.Fatal("expected value")
	}


	if value != "rohan" {
		t.Fatalf(
		  "expected rohan got %s",
		  value,
		)
	}
}

func TestTTLExpiration(t *testing.T){

	c := NewCache()


	c.Set(
		"temp",
		"value",
		100*time.Millisecond,
	)


	time.Sleep(
		200*time.Millisecond,
	)


	_, ok := c.Get("temp")


	if ok {
		t.Fatal("expected expired key")
	}
}