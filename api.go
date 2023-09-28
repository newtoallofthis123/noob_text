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
	if title == "" {
		title = "Untitled"
	}
	author := c.DefaultPostForm("author", "Anonymous")
	content := c.PostForm("content")
	if content == "" {
		c.String(400, "Content cannot be empty")
		return
	}
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

func (api *APIServer) handleUpdatePage(c *gin.Context) {
	api.onlyAuth(c)
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
	if doc.Author != api.getUsername(c) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.HTML(http.StatusOK, "update.html", gin.H{
		"hash":       doc.Hash,
		"title":      doc.Title,
		"author":     doc.Author,
		"content":    doc.Content,
		"created_at": doc.CreatedAt,
	})
}

func (api *APIServer) handleUpdate(c *gin.Context) {
	api.onlyAuth(c)
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
	if doc.Author != api.getUsername(c) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	title := c.DefaultPostForm("title", "Untitled")
	if title == "" {
		title = "Untitled"
	}
	content := c.PostForm("content")
	if content == "" {
		c.String(400, "Content cannot be empty")
		return
	}

	req := &UpdateDocumentRequest{
		Hash:    hash,
		Title:   title,
		Content: content,
	}

	err = api.store.UpdateDoc(req)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.String(200, "Updated!")
}

func (api *APIServer) handleSearchPage(c *gin.Context) {
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

func (api *APIServer) handleDelete(c *gin.Context) {
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
	if doc.Author != api.getUsername(c) {
		c.Redirect(http.StatusFound, "/")
		return
	}

	err = api.store.DeleteDoc(hash)
	if err != nil {
		c.JSON(500, err)
		return
	}

	c.String(200, "Deleted!")
}

func (api *APIServer) handleUserDelete(c *gin.Context) {
	toDelete := c.Param("username")
	api.onlyAuth(c)
	username := api.getUsername(c)
	if username != toDelete {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You can only delete your own account",
		})
		return
	}

	err := api.store.DeleteUser(username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
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

	c.HTML(http.StatusOK, "search.html", gin.H{
		"results": results,
		"num":     len(results),
	})
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

func (api *APIServer) handleAuthSignout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.SetCookie("username", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/login")
}

func (api *APIServer) handleCheckAuth(c *gin.Context) {
	cookie, err := c.Cookie("token")
	if err != nil {
		c.String(http.StatusOK, "false")
		return
	}

	jwt, err := ValidateJWT(cookie)
	if err != nil {
		c.String(http.StatusOK, "false")
		return
	}

	if !jwt.Valid {
		c.String(http.StatusOK, "false")
		return
	}

	c.String(http.StatusOK, "true")
}

func (api *APIServer) getUsername(c *gin.Context) string {
	username, err := c.Cookie("username")
	if err != nil {
		c.Redirect(http.StatusFound, "/authsignout")
		return ""
	}
	if username == "" {
		c.Redirect(http.StatusFound, "/authsignout")
		return ""
	}

	return username
}

func (api *APIServer) handleHome(c *gin.Context) {
	api.onlyAuth(c)

	var username string = api.getUsername(c)

	c.HTML(http.StatusOK, "index.html", gin.H{
		"username": username,
	})
}

func (api *APIServer) handleAbout(c *gin.Context) {
	c.HTML(http.StatusOK, "about.html", gin.H{})
}

func (api *APIServer) handleSource(c *gin.Context) {
	c.Redirect(http.StatusFound, "http://github.com/newtoallofthis123/noob_text")
}

func (api *APIServer) handleAccount(c *gin.Context) {
	api.onlyAuth(c)
	username := api.getUsername(c)

	docs, err := api.store.GetUserDocs(username)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.HTML(http.StatusOK, "account.html", gin.H{
		"username": username,
		"docs":     docs,
	})
}

func (api *APIServer) handleGetUserDocs(c *gin.Context) {
	username := c.Param("username")
	docs, err := api.store.GetUserDocs(username)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	c.HTML(http.StatusOK, "user.html", gin.H{
		"username": username,
		"docs":     docs,
	})
}

func (api *APIServer) onlyAuth(c *gin.Context) {
	cookie, err := c.Cookie("token")
	if err != nil {
		c.Redirect(http.StatusFound, "/about")
		return
	}
	if cookie == "" {
		c.Redirect(http.StatusFound, "/about")
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
	r.GET("/searchDoc", api.handleSearchPage)
	r.GET("/search", api.handleSearch)
	r.GET("/create", api.handleCreateForm)
	r.GET("/signup", api.handleSignupPage)
	r.GET("/login", api.handleLoginPage)
	r.GET("/signout", api.handleSignout)
	r.GET("/authsignout", api.handleAuthSignout)
	r.GET("/checkauth", api.handleCheckAuth)
	r.GET("/about", api.handleAbout)
	r.GET("/source", api.handleSource)
	r.GET("/account", api.handleAccount)
	r.GET("/user/:username", api.handleGetUserDocs)
	r.GET("/update/:hash", api.handleUpdatePage)

	r.POST("/create", api.handleCreateForm)
	r.POST("/createUser", api.handleSignup)
	r.POST("/loginUser", api.handleLogin)
	r.POST("/update_form/:hash", api.handleUpdate)

	r.DELETE("/delete/:hash", api.handleDelete)
	r.DELETE("/deleteUser/:username", api.handleUserDelete)

	err = r.Run(api.listenAddr)
	return err
}
