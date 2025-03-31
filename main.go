package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	initDb()
	initDir()
	initSecret()
	run()
}

func run() {
	// gin.SetMode(gin.ReleaseMode)
	// r := gin.New()
	r := gin.Default()
	r.Use(cors())

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

	v1 := r.Group("api/v1", auth)
	v1.GET("/ws/message", messageHandler)
	v1.GET("/ws/convid", convIdHandler)
	v1.POST("/upload", upFileHandler)
	v1.GET("/files/:filename", downFileHandler)

	user := v1.Group("user")
	user.GET("/me", userOwnInfo)
	user.GET("/list", userList)
	user.POST("/rename", resetName)
	user.POST("/repassword", resetPassword)

	group := v1.Group("group")
	group.GET("/list", groupListAsMember)
	group.POST("/create", groupCreate)
	group.POST("/info", groupInfo)
	group.POST("/join", groupJoin)
	group.POST("/member", memberList)
	group.POST("/leave", memberLeave)
	group.POST("/remove", memberRemove)

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

const (
	s = "success"
	e = "error"
)

var jwtSecret []byte

type resp[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type claims struct {
	Id int `json:"id"`
	jwt.RegisteredClaims
}

func cors() gin.HandlerFunc {
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

func auth(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, resp[any]{
			Message: "token invalid",
		})
		return
	}

	claim, err := parseToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, resp[any]{
			Message: err.Error(),
		})
		return
	}

	c.Set("userId", claim.Id)
	c.Next()
}

func checkPassword(password, hash string) bool {
	return hashPassword(password) == hash
}

func validateUserId(uid int) bool {
	return uid >= 13000000000 && uid <= 19999999999
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func getObjInfo[T *User | *Group](model T) error {
	return db.Model(model).First(model).Error
}

func initSecret() {
	var secret Secret
	db.Find(&secret)
	if secret.Key == "" {
		secret.Key = generateSecret()
		db.Create(secret)
	}
	jwtSecret = []byte(secret.Key)
}

func generateSecret() string {
	key := make([]byte, 8)
	rand.Read(key)
	return hex.EncodeToString(key[:])
}

func generateToken(uid int) (string, error) {
	claim := claims{
		Id: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(jwtSecret)
}

func parseToken(tokenString string) (*claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, errors.New("token invalid")
	}

	if claim, ok := token.Claims.(*claims); ok && token.Valid {
		return claim, nil
	}

	return nil, errors.New(e)
}

func validateGroupMember(gid, uid int) error {
	var count int64
	err := db.Model(&GroupMember{}).Where("group_id = ? AND user_id = ?", gid, uid).Count(&count).Error
	if err != nil {
		return errors.New(e)
	}

	if count == 0 {
		return errors.New("group invalid")
	}

	return nil
}

func validateGroup(gid int) error {
	var count int64
	err := db.Model(&Group{}).Where("id = ?", gid).Count(&count).Error
	if err != nil {
		return errors.New(e)
	}

	if count == 0 {
		return errors.New("group invalid")
	}

	return nil
}
