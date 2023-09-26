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
}

type DBInstance struct {
	db *sql.DB
}

func (pq *DBInstance) createTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS content(
			hash TEXT UNIQUE PRIMARY KEY,
			author TEXT NOT NULL DEFAULT 'Anonymous', 
			title TEXT NOT NULL DEFAULT 'Untitled',
			content TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`

	_, err := pq.db.Exec(query)
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
