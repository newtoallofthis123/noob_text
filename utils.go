package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/joho/godotenv"
)

func GetEnv() *Env {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	env := &Env{
		User:     os.Getenv("USER"),
		Password: os.Getenv("PASSWORD"),
		DB:       os.Getenv("DB"),
		PORT:     os.Getenv("PORT"),
		URL:      os.Getenv("URL"),
	}

	return env
}

func RanHash() string {
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var slug string

	for i := 0; i < 8; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			fmt.Println("Error generating random index:", err)
			return ""
		}
		slug += string(characters[randomIndex.Int64()])
	}

	return slug
}
