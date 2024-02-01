package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/newtoallofthis123/noob_text/utils"
)

type Store interface {
	createTable() error
	InsertIntoDB(req *utils.CreateDocumentRequest) error
	GetByHash(hash string) (utils.Document, error)
	GetAll() ([]utils.Document, error)
	CreateUser(req utils.CreateUserRequest) error
	GetUser(username string) (utils.User, error)
	GetUserDocs(username string) ([]utils.Document, error)
	UpdateDoc(req *utils.UpdateDocumentRequest) error
	DeleteDoc(hash string) error
	DeleteUser(username string) error
}

type DBInstance struct {
	db *sql.DB
}

func (pq *DBInstance) createTable() error {
	user_query := `
		CREATE TABLE IF NOT EXISTS users(
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`

	_, err := pq.db.Exec(user_query)

	if err != nil {
		return err
	}

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

	_, err = pq.db.Exec(query)

	return err
}

func NewStoreInstance() (*DBInstance, error) {
	env := utils.GetEnv()
	connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%s sslmode=disable", env.User, env.Password, env.URL, env.DB, env.PORT)
	fmt.Println(connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &DBInstance{db}, nil
}

func (pq *DBInstance) InsertIntoDB(req *utils.CreateDocumentRequest) error {
	query := `
		INSERT INTO content (hash, author, title, content)
		VALUES ($1, $2, $3, $4)
	`

	_, err := pq.db.Exec(query, req.Hash, req.Author, req.Title, req.Content)
	return err
}

func (pq *DBInstance) GetByHash(hash string) (utils.Document, error) {
	query := `
	SELECT * from content WHERE hash=$1
	`

	var document utils.Document

	rows := pq.db.QueryRow(query, hash)
	err := rows.Scan(&document.Hash, &document.Author, &document.Title, &document.Content, &document.CreatedAt, &document.UpdatedAt)
	if err != nil {
		return document, err
	}

	return document, nil
}

func (pq *DBInstance) GetAll() ([]utils.Document, error) {
	query := `
	SELECT * from content
	`

	var documents []utils.Document

	rows, err := pq.db.Query(query)
	if err != nil {
		return documents, err
	}

	for rows.Next() {
		var document utils.Document
		err := rows.Scan(&document.Hash, &document.Author, &document.Title, &document.Content, &document.CreatedAt, &document.UpdatedAt)
		if err != nil {
			return documents, err
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (pq *DBInstance) CreateUser(req utils.CreateUserRequest) error {
	query := `
	INSERT INTO users (username, password)
	VALUES ($1, $2)
	`

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	_, err = pq.db.Exec(query, req.Username, hashedPassword)
	return err
}

func (pq *DBInstance) GetUser(username string) (utils.User, error) {
	query := `
	SELECT * from users WHERE username=$1
	`

	var user utils.User

	rows := pq.db.QueryRow(query, username)
	err := rows.Scan(&user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return user, err
	}

	return user, nil
}

func (pq *DBInstance) GetUserDocs(username string) ([]utils.Document, error) {
	query := `
	SELECT * from content WHERE author=$1
	`

	var documents []utils.Document

	rows, err := pq.db.Query(query, username)
	if err != nil {
		return documents, err
	}

	for rows.Next() {
		var document utils.Document
		err := rows.Scan(&document.Hash, &document.Author, &document.Title, &document.Content, &document.CreatedAt, &document.UpdatedAt)
		if err != nil {
			return documents, err
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func (pq *DBInstance) UpdateDoc(req *utils.UpdateDocumentRequest) error {
	query := `
	UPDATE content SET title=$1, content=$2, updated_at=NOW() WHERE hash=$3
	`

	_, err := pq.db.Exec(query, req.Title, req.Content, req.Hash)
	return err
}

func (pq *DBInstance) DeleteDoc(hash string) error {
	query := `
	DELETE FROM content WHERE hash=$1
	`

	_, err := pq.db.Exec(query, hash)
	return err
}

func (pq *DBInstance) DeleteUser(username string) error {
	query := `
	DELETE FROM users WHERE username=$1
	`

	_, err := pq.db.Exec(query, username)
	return err
}
