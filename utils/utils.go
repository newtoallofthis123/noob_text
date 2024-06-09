package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func GetEnv() *Env {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	env := &Env{
		DatabaseUrl:   os.Getenv("DATABASE_URL"),
		RedisURL:      os.Getenv("REDIS_URL"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisDB:       os.Getenv("REDIS_DB"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		JwtSecret:     os.Getenv("JWT_SECRET"),
		Port:          os.Getenv("PORT"),
	}

	return env
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func MatchPasswords(toCheck string, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(toCheck))
	return err == nil
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

func CreateJWT(username string) (string, error) {
	env := GetEnv()
	claims := &jwt.MapClaims{
		// expire in like a week or two
		"expiresAt": time.Hour * 24 * 7 * 2,
		"username":  username,
	}
	secret := env.JwtSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	env := GetEnv()
	secret := env.JwtSecret

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}
