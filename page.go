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
	if checkSession(c, "/login", false) {
		return
	}
	c.FileFromFS("web/", http.FS(web))
}

func loginPage(c *gin.Context) {
	if checkSession(c, "/", true) {
		return
	}
	c.FileFromFS("web/login.html", http.FS(web))
}

func registerPage(c *gin.Context) {
	if checkSession(c, "/", true) {
		return
	}
	c.FileFromFS("web/register.html", http.FS(web))
}

func checkSession(c *gin.Context, redirectPath string, should bool) bool {
	cookie, err := c.Cookie("session_id")
	valid := false
	if err == nil {
		_, valid = validateSession(cookie)
	}

	if valid == should {
		c.Redirect(http.StatusFound, redirectPath)
		return true
	}
	return false
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
