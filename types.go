package main

type Env struct {
	User     string
	Password string
	DB       string
	PORT     string
	URL      string
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
