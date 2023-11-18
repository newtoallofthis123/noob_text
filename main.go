package main

import "fmt"

func main() {

	store, err := NewStoreInstance()
	if err != nil {
		panic(err)
	}

	err = store.createTable()
	if err != nil {
		panic(err)
	}

	cache, err := NewRedisClient()
	if err != nil {
		panic(err)
	}

	//testing connection
	cache.client.Set(cache.ctx, "key", "value", 0)

	val, err := cache.client.Get(cache.ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	api := APIServer{
		listenAddr: ":3579",
		store:      store,
		cache:      cache,
	}

	fmt.Println("Starting API server")
	err = api.Start()
	if err != nil {
		panic(err)
	}
}
