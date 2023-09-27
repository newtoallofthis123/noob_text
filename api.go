package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	listenAddr string
	store      Store
}

func (api *APIServer) handleGetByHash(c *gin.Context) {
	hash := c.Param("hash")
	doc, err := api.store.GetByHash(hash)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(404, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(500, err)
		return
	}

	c.HTML(http.StatusOK, "doc.html", gin.H{
		"title":      doc.Title,
		"author":     doc.Author,
		"content":    doc.Content,
		"created_at": doc.CreatedAt,
	})
}

func (api *APIServer) handleCreateForm(c *gin.Context) {
	title := c.DefaultPostForm("title", "Untitled")
	author := c.DefaultPostForm("author", "Anonymous")
	content := c.PostForm("content")
	hash := RanHash()

	req := &CreateDocumentRequest{
		Hash:    hash,
		Title:   title,
		Author:  author,
		Content: content,
	}

	err := api.store.InsertIntoDB(req)
	if err != nil {
		c.String(500, err.Error())
		return
	}
	c.String(200, "<a class='text-blue-600' href='/docs/"+hash+"'>/docs/"+hash+"</a>")
}

func (api *APIServer) handleSearch(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	docs, err := api.store.GetAll()
	if err != nil {
		c.JSON(500, err)
		return
	}

	var results []Document

	for _, patient := range docs {
		docString := fmt.Sprintf("%v", patient)
		//remove all the -, :, _ and lowercase the string
		docString = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(docString, "-", ""), ":", ""), "{", ""))
		if strings.Contains(docString, query) {
			results = append(results, patient)
		}
	}

	c.JSON(200, results)
}

func (api *APIServer) handleSignup(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	err := api.store.CreateUser(CreateUserRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		log.Default().Println(err)
		c.JSON(500, err)
		return
	}

	user, err := api.store.GetUser(username)
	if err != nil {
		log.Default().Println(err)
		c.JSON(500, err)
		return
	}

	jwtToken, err := CreateJWT(user.Username)
	if err != nil {
		log.Default().Println(err)
		c.JSON(500, err)
		return
	}

	c.SetCookie("token", jwtToken, 7*24*60*60, "/", "localhost", false, true)
	c.String(http.StatusOK, "User created")
}

func (api *APIServer) handleSignupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "signup.html", gin.H{})
}

func (api *APIServer) handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := api.store.GetUser(username)
	if err != nil {
		log.Default().Println(err)
		c.String(http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !MatchPasswords(password, user.Password) {
		c.String(http.StatusUnauthorized, "Invalid credentials")
		return
	}

	jwtToken, err := CreateJWT(user.Username)
	if err != nil {
		log.Default().Println(err)
		c.String(http.StatusInternalServerError, "Error creating JWT")
		return
	}

	c.SetCookie("username", user.Username, 7*24*60*60, "/", "localhost", false, true)
	c.SetCookie("token", jwtToken, 7*24*60*60, "/", "localhost", false, true)
	c.String(http.StatusOK, "User Logged in")
}

func (api *APIServer) handleLoginPage(c *gin.Context) {
	cookie, err := c.Cookie("token")
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}
	if cookie != "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	jwt, err := ValidateJWT(cookie)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}
	if !jwt.Valid {
		c.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func (api *APIServer) handleSignout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
}

func (api *APIServer) handleHome(c *gin.Context) {
	api.onlyAuth(c)
	username, err := c.Cookie("username")
	if err != nil {
		c.Redirect(http.StatusFound, "/authsignout")
		return
	}
	if username == "" {
		c.Redirect(http.StatusFound, "/authsignout")
		return
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"username": username,
	})
}

func (api *APIServer) onlyAuth(c *gin.Context) {
	cookie, err := c.Cookie("token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if cookie == "" {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	jwt, err := ValidateJWT(cookie)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !jwt.Valid {
		c.Redirect(http.StatusFound, "/login")
		return
	}
}

func (api *APIServer) Start() error {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./public")

	err := r.SetTrustedProxies(nil)
	if err != nil {
		return err
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, "pong")
	})
	r.GET("/", api.handleHome)
	r.GET("/docs/:hash", api.handleGetByHash)
	r.GET("/search", api.handleSearch)
	r.GET("/signup", api.handleSignupPage)
	r.GET("/login", api.handleLoginPage)
	r.GET("/signout", api.handleSignout)

	r.POST("/create", api.handleCreateForm)
	r.POST("/createUser", api.handleSignup)
	r.POST("/loginUser", api.handleLogin)

	err = r.Run(api.listenAddr)
	return err
}
