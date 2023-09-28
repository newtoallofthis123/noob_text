package main

type Env struct {
	User      string
	Password  string
	DB        string
	PORT      string
	URL       string
	JwtSecret string
}

type CreateDocumentRequest struct {
	Hash    string
	Author  string
	Title   string
	Content string
}

type Document struct {
	Hash      string
	Author    string
	Title     string
	Content   string
	CreatedAt string
	UpdatedAt string
}

type CreateUserRequest struct {
	Username string
	Password string
}

type User struct {
	Username  string
	Password  string
	CreatedAt string
}

type UpdateDocumentRequest struct {
	Hash    string
	Title   string
	Content string
}
