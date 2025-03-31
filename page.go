package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:web
var web embed.FS

func indexPage(c *gin.Context) {
	if !check(c) {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.FileFromFS("web/", http.FS(web))
}

func loginPage(c *gin.Context) {
	if check(c) {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.FileFromFS("web/login.html", http.FS(web))
}

func registerPage(c *gin.Context) {
	if check(c) {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.FileFromFS("web/register.html", http.FS(web))
}

func check(c *gin.Context) bool {
	token := c.GetHeader("Authorization")
	if token == "" {
		return false
	}

	_, err := parseToken(token)
	if err != nil {
		return false
	}

	return true
}

func favicon(c *gin.Context) {
	c.FileFromFS("web/favicon.ico", http.FS(web))
}

func staticFileHandler(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file := c.Param("file")
		c.FileFromFS(fmt.Sprintf("web/%s/%s", prefix, file), http.FS(web))
	}
}

func cacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=3600")
		c.Next()
	}
}
