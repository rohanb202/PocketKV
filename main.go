package main

import (
	"fmt"
	"sync"
	"time"
	"container/heap"
)

type Item struct {
	Value string
	Expiry time.Time
}

type ExpiryItem struct {
	key string
	expiry time.Time
	index int
}

type Cache struct {
	data map[string]Item
	mu   sync.RWMutex
	expiryHeap *ExpiryHeap
}


type ExpiryHeap []*ExpiryItem


func (h ExpiryHeap) Len() int {
	return len(h)
}


func (h ExpiryHeap) Less(i,j int) bool {

	return h[i].expiry.Before(h[j].expiry)

}


func (h ExpiryHeap) Swap(i,j int){

	h[i],h[j] = h[j],h[i]

}


func (h *ExpiryHeap) Push(x any){

	*h = append(*h,x.(*ExpiryItem))

}


func (h *ExpiryHeap) Pop() any {

	old := *h

	n := len(old)

	item := old[n-1]

	*h = old[:n-1]

	return item
}







func NewCache() *Cache {

	h:= make(ExpiryHeap, 0)
	heap.Init(&h)

	return &Cache{
		data: make(map[string]Item),
		expiryHeap: &h,
	}
}
func (c *Cache) Set(key, value string, expiry time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = Item{Value: value, Expiry: expiry}
	heap.Push(c.expiryHeap, &ExpiryItem{key: key, expiry: expiry})
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	
	value, ok := c.data[key]

	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if value.Expiry.Before(time.Now()) {
		delete(c.data, key)
		return "", false
	}
	return value.Value, ok
}
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for c.expiryHeap.Len() > 0 {
		item := (*c.expiryHeap)[0]
		if item.expiry.After(now) {
			break
		}
		current, exists := c.data[item.key]
		if exists && current.Expiry.Equal(item.expiry) {
			delete(c.data, item.key)
		}
		heap.Pop(c.expiryHeap)
	}
}

func (c *Cache) StartCleanup(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			c.Cleanup()
		}
	}()
}

func main(){
    cache := NewCache()
	cache.StartCleanup(1 * time.Minute) // Start cleanup every minute

	cache.Set(
		"name",
		"rohan",
		time.Now().Add(5*time.Minute),
    )

	value, exists := cache.Get("name")


	if exists {
		fmt.Println("Value:", value)
	}


	cache.Delete("name")


	_, exists = cache.Get("name")

	fmt.Println("Exists after delete:", exists)
}