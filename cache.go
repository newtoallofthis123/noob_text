package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	CreateUser(user User) error
	CreateDoc(doc Document) error
	GetUser(username string) (User, error)
	GetDoc(hash string) (Document, error)
	GetAllDocs() ([]Document, error)
	DeleteUser(username string)
	DeleteDoc(hash string)
}

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewClient returns a new redis client
func NewRedisClient() (*RedisClient, error) {
	env := GetEnv()
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

func (r *RedisClient) CreateUser(user User) error {
	json_encoded, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// convert the marshalled json to a strin
	r.client.Set(r.ctx, user.Username, string(json_encoded), time.Second*3600)
	return nil
}

func (r *RedisClient) CreateDoc(doc Document) error {
	json_encoded, err := json.Marshal(doc)
	if err != nil {
		fmt.Println(err)
		return err
	}
	r.client.Set(r.ctx, doc.Hash, string(json_encoded), time.Second*3600)
	return nil
}

func (r *RedisClient) GetUser(username string) (User, error) {
	var user User
	json_string, err := r.client.Get(r.ctx, username).Result()
	if err != nil {
		return User{}, err
	}
	json.Unmarshal([]byte(json_string), &user)
	if user.Username == "" {
		return User{}, fmt.Errorf("User not found")
	}
	return user, nil
}

func (r *RedisClient) GetDoc(hash string) (Document, error) {
	var doc Document
	json_string, err := r.client.Get(r.ctx, hash).Result()
	if err != nil {
		return Document{}, err
	}
	json.Unmarshal([]byte(json_string), &doc)
	if doc.Hash == "" {
		return Document{}, fmt.Errorf("Document not found")
	}
	return doc, nil
}

func (r *RedisClient) GetAllDocs() ([]Document, error) {
	var docs []Document
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
