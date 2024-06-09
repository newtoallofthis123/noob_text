package main

import (
	"fmt"

	"github.com/newtoallofthis123/noob_text/utils"
)

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
	env := utils.GetEnv()

	api := APIServer{
		listenAddr: ":" + env.Port,
		store:      store,
		cache:      cache,
	}

	fmt.Println("Starting API server")
	err = api.Start()
	if err != nil {
		panic(err)
	}
}
