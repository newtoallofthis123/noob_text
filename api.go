package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/newtoallofthis123/noob_text/templates"
	"github.com/newtoallofthis123/noob_text/utils"
)

type APIServer struct {
	listenAddr string
	store      Store
	cache      Cache
}

func (api *APIServer) handleGetByHash(c *gin.Context) {
	hash := c.Param("hash")

	doc, err := api.cache.GetDoc(hash)
	fmt.Println("Getting using hash")
	if err != nil {
		fmt.Println(err)
	} else {
		templates.Base(doc.Title, templates.Doc(doc)).Render(c.Request.Context(), c.Writer)
		return
	}

	doc, err = api.store.GetByHash(hash)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(404, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(500, err)
		return
	}

	err = api.cache.CreateDoc(doc)
	if err != nil {
		fmt.Println(err)
	}

	templates.Base(doc.Title, templates.Doc(doc)).Render(c.Request.Context(), c.Writer)
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
	hash := utils.RanHash()

	req := &utils.CreateDocumentRequest{
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
	// we choose not to cache a new document because it is unlikely to be accessed again
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

	templates.Base("Update", templates.Update(
		doc.Title,
		doc.Content,
		doc.Author,
		doc.CreatedAt,
		hash,
	)).Render(c.Request.Context(), c.Writer)
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

	//delete the old document from cache
	api.cache.DeleteDoc(hash)

	req := &utils.UpdateDocumentRequest{
		Hash:    hash,
		Title:   title,
		Content: content,
	}

	err = api.store.UpdateDoc(req)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	// re-caching the doc requires a new query to the database
	// which is not ideal

	c.String(200, "Updated!")
}

func (api *APIServer) handleSearchPage(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	docs, err := api.store.GetAll()
	if err != nil {
		c.JSON(500, err)
		return
	}

	var results []utils.Document

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

	//delete the old document from cache
	api.cache.DeleteDoc(hash)

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

	//delete the old document from cache
	api.cache.DeleteUser(username)
}

func (api *APIServer) handleSearch(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	docs, err := api.store.GetAll()
	if err != nil {
		c.JSON(500, err)
		return
	}

	var results []utils.Document

	for _, doc := range docs {
		docString := fmt.Sprintf("%v", doc)
		//remove all the -, :, _ and lowercase the string
		docString = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(docString, "-", ""), ":", ""), "{", ""))
		if strings.Contains(docString, query) {
			results = append(results, doc)
		}
	}

	templates.Base("Search Results", templates.Search(results, len(results))).Render(c.Request.Context(), c.Writer)
}

func (api *APIServer) handleSignup(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	err := api.store.CreateUser(utils.CreateUserRequest{
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

	jwtToken, err := utils.CreateJWT(user.Username)
	if err != nil {
		log.Default().Println(err)
		c.JSON(500, err)
		return
	}

	c.SetCookie("token", jwtToken, 7*24*60*60, "/", "localhost", false, true)
	c.String(http.StatusOK, "User created")
}

func (api *APIServer) handleSignupPage(c *gin.Context) {
	templates.Base("Signup", templates.SignUp()).Render(c.Request.Context(), c.Writer)
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

	if !utils.MatchPasswords(password, user.Password) {
		c.String(http.StatusUnauthorized, "Invalid credentials")
		return
	}

	jwtToken, err := utils.CreateJWT(user.Username)
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
		templates.Base("Login", templates.Login()).Render(c.Request.Context(), c.Writer)
		return
	}
	if cookie != "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	jwt, err := utils.ValidateJWT(cookie)
	if err != nil || !jwt.Valid {
		templates.Base("Login", templates.Login()).Render(c.Request.Context(), c.Writer)
		return
	}

	templates.Base("Login", templates.Login()).Render(c.Request.Context(), c.Writer)
}

func (api *APIServer) handleSignout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.SetCookie("username", "", -1, "/", "localhost", false, true)
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

	jwt, err := utils.ValidateJWT(cookie)
	if err != nil || !jwt.Valid {
		c.String(http.StatusOK, "false")
		return
	}

	c.String(http.StatusOK, "true")
}

func (api *APIServer) getUsername(c *gin.Context) string {
	username, err := c.Cookie("username")
	if err != nil || username == "" {
		c.Redirect(http.StatusFound, "/authsignout")
		return ""
	}

	return username
}

func (api *APIServer) handleHome(c *gin.Context) {
	api.onlyAuth(c)

	var username string = api.getUsername(c)

	templates.Base("NoobText", templates.IndexPage(username)).Render(c.Request.Context(), c.Writer)
}

func (api *APIServer) handleAbout(c *gin.Context) {
	templates.Base("About NoobText", templates.AboutPage()).Render(c.Request.Context(), c.Writer)
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

	templates.Base("Account", templates.Account(username, docs)).Render(c.Request.Context(), c.Writer)
}

// handleGetUserDocs is a handler for the GET /user/:username route which
// renders a list of documents for a given user.

func (api *APIServer) handleGetUserDocs(c *gin.Context) {
	username := c.Param("username")
	docs, err := api.store.GetUserDocs(username)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	templates.Base("User Docs", templates.User(username, docs)).Render(c.Request.Context(), c.Writer)
}

func (api *APIServer) onlyAuth(c *gin.Context) {
	cookie, err := c.Cookie("token")
	if err != nil || cookie == "" {
		c.Redirect(http.StatusFound, "/about")
		return
	}

	jwt, err := utils.ValidateJWT(cookie)
	if err != nil || !jwt.Valid {
		c.Redirect(http.StatusFound, "/login")
		return
	}
}

func (api *APIServer) Start() error {

	// use log file as gin's output
	// check if logs folder exists
	// if _, err := os.Stat("logs"); os.IsNotExist(err) {
	// 	os.Mkdir("logs", 0755)
	// }

	// _, err := os.Create(fmt.Sprintf("logs/%s.log", get_date()))
	// if err != nil {
	// 	return err
	// }

	// f, err := os.OpenFile(fmt.Sprintf("logs/%s.log", get_date()), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	return err
	// }

	r := gin.Default()
	// gin.DefaultWriter = f

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

	fmt.Println("Listening on", api.listenAddr)

	err = r.Run(api.listenAddr)
	return err
}
