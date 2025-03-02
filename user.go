package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// post /login
func userLogin(c *gin.Context) {
	var credentials struct {
		ID       int    `json:"id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式无效",
		})
		return
	}

	var user User
	err = db.Select("id", "password").First(&user, "id = ?", credentials.ID).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "无效凭证",
		})
		return
	}

	if !checkPassword(credentials.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "无效凭证",
		})
		return
	}

	sessionId, err := createSession(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "登陆失败",
		})
		return
	}

	c.SetCookie(
		"session_id", sessionId,
		72*3600,
		"/",
		"",
		c.Request.URL.Scheme == "https",
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"msg": "登陆成功",
	})
}

// post /register
func userRegister(c *gin.Context) {
	var newUser struct {
		ID       int    `json:"id" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&newUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式无效",
		})
		return
	}

	if !validateUserId(newUser.ID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户 ID 格式",
		})
		return
	}

	hashedPassword := hashPassword(newUser.Password)

	user := User{
		ID:       newUser.ID,
		Name:     strings.TrimSpace(newUser.Name),
		Password: hashedPassword,
	}

	err = db.Create(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "用户已存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "注册失败",
			})
		}
		return
	}

	sessionId, err := createSession(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建 Cookie 失败",
		})
		return
	}

	c.SetCookie(
		"session_id",
		sessionId,
		72*3600,
		"/",
		"",
		c.Request.URL.Scheme == "https",
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"msg": "注册成功",
	})
}

// get /api/v1/user/info/me
func userOwnInfo(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	userName, err := getNameById(&User{}, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "用户不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "查找信息失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   userId,
		"name": userName,
	})
}

// get /api/v1/user/lists
func userList(c *gin.Context) {
	var users []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	err := db.Model(&User{}).
		Select("id", "name").
		Order("id DESC").
		Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "检索用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

// get /api/v1/user/logout
func userLogout(c *gin.Context) {
	c.SetCookie(
		"session_id",
		"",
		-1,
		"/",
		"",
		c.Request.URL.Scheme == "https",
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"msg": "退出登陆成功",
	})
}

// post /api/v1/user/rename
func resetName(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var request struct {
		Name string `json:"name" binding:"required"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的用户名格式",
		})
		return
	}

	result := db.Model(&User{}).
		Where("id = ?", userId).
		Update("name", strings.TrimSpace(request.Name))

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "修改失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "修改用户名成功",
	})
}

// post /api/v1/user/repassword
func resetPassword(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var request struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式无效",
		})
		return
	}

	var user User
	err = db.Select("password").First(&user, "id = ?", userId).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户 ID 错误",
		})
		return
	}

	if !checkPassword(request.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "旧密码错误",
		})
		return
	}

	newHashedPassword := hashPassword(request.NewPassword)

	result := db.Model(&User{}).
		Where("id = ?", userId).
		Updates(map[string]interface{}{
			"password": newHashedPassword,
		})

	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "修改密码失败",
		})
		return
	}

	db.Delete(&Session{}, "user_id = ?", userId)

	c.JSON(http.StatusOK, gin.H{
		"msg": "修改密码成功",
	})
}

func validateUserId(id int) bool {
	return id >= 13000000000 && id <= 19999999999
}

func hashPassword(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}

func checkPassword(input, storedHash string) bool {
	return hashPassword(input) == storedHash
}

func createSession(userId int) (string, error) {
	session := Session{
		ID:        uuid.New().String(),
		UserId:    userId,
		ExpiresAt: time.Now().Add(72 * time.Hour),
	}

	err := db.Create(&session).Error
	if err != nil {
		return "", err
	}

	return session.ID, nil
}

func validateSession(sessionId string) (int, bool) {
	var session Session
	err := db.Where("id = ?", sessionId).First(&session).Error
	if err != nil {
		return 0, false
	}

	if time.Now().After(session.ExpiresAt) {
		return 0, false
	}

	return session.UserId, true
}

func getNameById(model interface{}, id int) (name string, err error) {
	err = db.Select("name").
		First(model, "id = ?", id).
		Scan(&name).Error
	return name, err
}
