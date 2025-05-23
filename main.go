package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	initDb()
	initDir()
	run()
}

func run() {
	// gin.SetMode(gin.ReleaseMode)
	// r := gin.New()
	r := gin.Default()
	r.Use(Cors())

	routesAdmin(r)

	page := r.Group("/")
	page.Use(cacheControl())
	page.GET("/", indexPage)
	page.GET("/login", loginPage)
	page.GET("/register", registerPage)
	page.GET("/favicon.ico", favicon)
	page.GET("/image/:file", staticFileHandler("image"))
	page.GET("/js/*file", staticFileHandler("js"))
	page.GET("/css/*file", staticFileHandler("css"))

	r.POST("/login", userLogin)
	r.POST("/register", userRegister)

	v1 := r.Group("api/v1")
	v1.Use(AuthorizationMiddleware())
	v1.GET("/ws/message", messageHandler)
	v1.GET("/ws/convid", convIdHandler)
	v1.POST("/upload", upFileHandler)
	v1.GET("/files/:filename", downFileHandler)

	user := v1.Group("user")
	user.GET("/info/me", userOwnInfo)
	user.GET("/lists", userList)
	user.GET("/logout", userLogout)
	user.POST("/rename", resetName)
	user.POST("/repassword", resetPassword)

	group := v1.Group("group")
	group.GET("/lists", groupListAsMember)
	group.GET("/member/:id", groupMember)
	group.GET("/info/:id", groupInfo)
	group.POST("/create", groupCreate)
	group.GET("/join/:id", memberJoin)
	group.GET("/leave/:id", memberLeave)
	group.GET("/remove/:gid/:mid", memberRemove)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("[err]", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("[err]", err)
	}

	_db, err := db.DB()
	if err != nil {
		log.Fatal("[err]", err)
	}
	err = _db.Close()
	if err != nil {
		log.Fatal("[err]", err)
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
