package main

import (
	"net/http"

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

func (api *APIServer) handleHome(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
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
	r.POST("/create", api.handleCreateForm)

	err = r.Run(api.listenAddr)
	return err
}
