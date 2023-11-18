package main

type Env struct {
	User          string
	Password      string
	DB            string
	PORT          string
	URL           string
	JwtSecret     string
	RedisURL      string
	RedisPort     string
	RedisDB       string
	RedisPassword string
}

type CreateDocumentRequest struct {
	Hash    string
	Author  string
	Title   string
	Content string
}

type Document struct {
	Hash      string `json:"hash"`
	Author    string `json:"author"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CreateUserRequest struct {
	Username string
	Password string
}

// Introducing the JSON tag for caching
type User struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

type UpdateDocumentRequest struct {
	Hash    string
	Title   string
	Content string
}
