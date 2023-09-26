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

	api := APIServer{
		listenAddr: "localhost:3579",
		store:      store,
	}

	fmt.Println("Starting API server")
	err = api.Start()
	if err != nil {
		panic(err)
	}
}
