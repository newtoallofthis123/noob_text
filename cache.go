package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/newtoallofthis123/noob_text/utils"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	CreateUser(user utils.User) error
	CreateDoc(doc utils.Document) error
	GetUser(username string) (utils.User, error)
	GetDoc(hash string) (utils.Document, error)
	GetAllDocs() ([]utils.Document, error)
	DeleteUser(username string)
	DeleteDoc(hash string)
}

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewClient returns a new redis client
func NewRedisClient() (*RedisClient, error) {
	env := utils.GetEnv()
	db, err := strconv.Atoi(env.RedisDB)
	if err != nil {
		fmt.Println("The redis db is not a number")
	}
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     env.RedisURL + ":" + env.RedisPort,
			Password: "",
			DB:       db,
		}),
		ctx: context.Background(),
	}, nil
}

func (r *RedisClient) CreateUser(user utils.User) error {
	json_encoded, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// convert the marshalled json to a string
	r.client.Set(r.ctx, user.Username, string(json_encoded), time.Second*3600)
	return nil
}

func (r *RedisClient) CreateDoc(doc utils.Document) error {
	json_encoded, err := json.Marshal(doc)
	if err != nil {
		fmt.Println(err)
		return err
	}
	r.client.Set(r.ctx, doc.Hash, string(json_encoded), time.Second*3600)
	return nil
}

func (r *RedisClient) GetUser(username string) (utils.User, error) {
	var user utils.User
	json_string, err := r.client.Get(r.ctx, username).Result()
	if err != nil {
		return utils.User{}, err
	}
	json.Unmarshal([]byte(json_string), &user)
	if user.Username == "" {
		return utils.User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

func (r *RedisClient) GetDoc(hash string) (utils.Document, error) {
	var doc utils.Document
	json_string, err := r.client.Get(r.ctx, hash).Result()
	if err != nil {
		return utils.Document{}, err
	}
	json.Unmarshal([]byte(json_string), &doc)
	if doc.Hash == "" {
		return utils.Document{}, fmt.Errorf("document not found")
	}
	return doc, nil
}

func (r *RedisClient) GetAllDocs() ([]utils.Document, error) {
	var docs []utils.Document
	keys, err := r.client.Keys(r.ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		doc, err := r.GetDoc(key)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (r *RedisClient) DeleteUser(username string) {
	r.client.Del(r.ctx, username)
}

func (r *RedisClient) DeleteDoc(hash string) {
	r.client.Del(r.ctx, hash)
}
