package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	initDb()
	initDir()
	run()
}

func run() {
	r := gin.Default()
	r.Use(Cors())

	r.GET("/", indexPage)
	r.GET("/login", loginPage)
	r.GET("/register", registerPage)
	r.GET("/favicon.ico", getFavicon)
	r.GET("/image/:file", staticFileHandler("image"))
	r.GET("/js/*file", staticFileHandler("js"))
	r.GET("/css/:file", staticFileHandler("css"))
	r.POST("/login", userLogin)
	r.POST("/register", userRegister)

	v1 := r.Group("api/v1")
	v1.Use(AuthorizationMiddleware())
	v1.GET("/ws/message", messageHandler)
	v1.GET("/ws/convid", convIdHandler)
	v1.POST("/upload", upFileHandler)
	v1.GET("/files/:filename", downFileHandler)

	user := v1.Group("user")
	user.GET("/me", userOwnInfo)
	user.GET("/lists", userList)
	user.POST("/logout", userLogout)
	user.POST("/rename", resetName)
	user.POST("/repassword", resetPassword)

	group := v1.Group("group")
	group.GET("/lists", groupListAsMember)
	group.POST("/info", groupInfo)
	group.POST("/create", groupCreate)
	group.POST("/join", memberJoin)
	group.POST("/leave", memberLeave)

	if err := r.Run(":8080"); err != nil {
		panic("[err]" + err.Error())
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}

func AuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "需要认证",
			})
			return
		}

		userId, valid := validateSession(cookie)
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "无效或过期的 Cookie",
			})
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}
