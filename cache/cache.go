package cache

import (
	"context"
	"sync"
	"time"
	"container/heap"
)

type Item struct {
	Value string
	Expiry time.Time
	Version int64
	Deleted bool
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
	expiryMap map[string]*ExpiryItem
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
	h[i].index = i
	h[j].index = j

}


func (h *ExpiryHeap) Push(x any){

	item := x.(*ExpiryItem)

	item.index=len(*h)

	*h=append(*h,item)
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
		expiryMap: make(map[string]*ExpiryItem),
	}
}
func (c *Cache) Set(key, value string, ttl time.Duration, version int64) {
    c.mu.Lock()
    defer c.mu.Unlock()
    expiry := time.Now().Add(ttl)

	if existing, ok := c.data[key]; ok && existing.Version > version{
		return
	}

    c.data[key] = Item{Value: value, Expiry: expiry, Version: version}

    if existing, ok := c.expiryMap[key]; ok {
        existing.expiry = expiry
        heap.Fix(c.expiryHeap, existing.index)
    } else {
        expiryItem := &ExpiryItem{key: key, expiry: expiry}
        heap.Push(c.expiryHeap, expiryItem)
        c.expiryMap[key] = expiryItem
    }
}

func (c *Cache) Get(key string) (Item, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.data[key]
	if !ok {
		return Item{}, false
	}

	if !item.Deleted && item.Expiry.Before(time.Now()) {

		if expiryItem, ok := c.expiryMap[key]; ok {
			heap.Remove(c.expiryHeap, expiryItem.index)
			delete(c.expiryMap, key)
		}

		delete(c.data, key)

		return Item{}, false
	}

	return item, true 
}

func (c *Cache) Delete(
	key string,
	version int64,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.data[key]; ok &&
		existing.Version > version {
		return
	}

	if expiryItem, ok := c.expiryMap[key]; ok {
		heap.Remove(c.expiryHeap, expiryItem.index)
		delete(c.expiryMap, key)
	}

	c.data[key] = Item{
		Version: version,
		Deleted: true,
	}
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

func (c *Cache) StartCleanup(
	ctx context.Context,
	interval time.Duration,
){

	go func(){

		ticker:=time.NewTicker(interval)

		defer ticker.Stop()


		for {

			select {

			case <-ticker.C:
				c.Cleanup()


			case <-ctx.Done():
				return
			}
		}

	}()
}
