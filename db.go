package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Store interface {
	createTable() error
	InsertIntoDB(req *CreateDocumentRequest) error
	GetByHash(hash string) (Document, error)
	GetAll() ([]Document, error)
	CreateUser(req CreateUserRequest) error
	GetUser(username string) (User, error)
	GetUserDocs(username string) ([]Document, error)
}

type DBInstance struct {
	db *sql.DB
}

func (pq *DBInstance) createTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS content(
			hash TEXT UNIQUE PRIMARY KEY,
			author TEXT REFERENCES users(username), 
			title TEXT NOT NULL DEFAULT 'Untitled',
			content TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`

	_, err := pq.db.Exec(query)

	if err != nil {
		return err
	}

	user_query := `
		CREATE TABLE IF NOT EXISTS users(
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`

	_, err = pq.db.Exec(user_query)

	return err
}

func NewStoreInstance() (*DBInstance, error) {
	env := GetEnv()
	connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%s sslmode=disable", env.User, env.Password, env.URL, env.DB, env.PORT)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &DBInstance{db}, nil
}

func (pq *DBInstance) InsertIntoDB(req *CreateDocumentRequest) error {
	query := `
		INSERT INTO content (hash, author, title, content)
		VALUES ($1, $2, $3, $4)
	`

	_, err := pq.db.Exec(query, req.Hash, req.Author, req.Title, req.Content)
	return err
}

func (pq *DBInstance) GetByHash(hash string) (Document, error) {
	query := `
	SELECT * from content WHERE hash=$1
	`

	var document Document

	rows := pq.db.QueryRow(query, hash)
	err := rows.Scan(&document.Hash, &document.Author, &document.Title, &document.Content, &document.CreatedAt, &document.UpdatedAt)
	if err != nil {
		return document, err
	}

	return document, nil
}

func (pq *DBInstance) GetAll() ([]Document, error) {
	query := `
	SELECT * from content
	`

	var documents []Document

	rows, err := pq.db.Query(query)
	if err != nil {
		return documents, err
	}

	for rows.Next() {
		var document Document
		err := rows.Scan(&document.Hash, &document.Author, &document.Title, &document.Content, &document.CreatedAt, &document.UpdatedAt)
		if err != nil {
			return documents, err
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (pq *DBInstance) CreateUser(req CreateUserRequest) error {
	query := `
	INSERT INTO users (username, password)
	VALUES ($1, $2)
	`

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return err
	}

	_, err = pq.db.Exec(query, req.Username, hashedPassword)
	return err
}

func (pq *DBInstance) GetUser(username string) (User, error) {
	query := `
	SELECT * from users WHERE username=$1
	`

	var user User

	rows := pq.db.QueryRow(query, username)
	err := rows.Scan(&user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (pq *DBInstance) GetUserDocs(username string) ([]Document, error) {
	query := `
	SELECT * from content WHERE author=$1
	`

	var documents []Document

	rows, err := pq.db.Query(query, username)
	if err != nil {
		return documents, err
	}

	for rows.Next() {
		var document Document
		err := rows.Scan(&document.Hash, &document.Author, &document.Title, &document.Content, &document.CreatedAt, &document.UpdatedAt)
		if err != nil {
			return documents, err
		}

		documents = append(documents, document)
	}

	return documents, nil
}
