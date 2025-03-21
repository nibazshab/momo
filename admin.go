package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const defaultPassword = "111111"

var (
	timeout    time.Time
	adminToken = hashPassword(uuid.New().String())
)

func routesAdmin(r *gin.Engine) {
	r.GET("/admin", adminPage)
	r.POST("/admin/login", adminLogin)

	admin := r.Group("admin")
	admin.Use(timeoutMiddleware())
	admin.GET("/show_users", showAllUser)
	admin.GET("/show_groups", showAllGroup)
	admin.POST("/forget_password", forgetPassword)
	admin.POST("/delete_user", deleteUser)
	admin.POST("/delete_group", deleteGroup)

	log.Printf("超级管理员密码 %s", adminToken)
}

// get /admin
func adminPage(c *gin.Context) {
	if time.Now().After(timeout) {
		c.FileFromFS("web/admin/login.html", http.FS(web))
		return
	}
	c.FileFromFS("web/admin/", http.FS(web))
}

// post /admin/login
func adminLogin(c *gin.Context) {
	var token struct {
		Value string `json:"token" binding:"required"`
	}

	a := jsonData(&token, c)
	if !a {
		return
	}

	if adminToken != token.Value {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的超级管理员密码",
		})
		return
	}

	timeout = time.Now().Add(time.Hour)

	c.Status(http.StatusOK)
}

// get /admin/show_users
func showAllUser(c *gin.Context) {
	userList(c)
}

// get /admin/show_groups
func showAllGroup(c *gin.Context) {
	var groups []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		OwnerId int    `json:"owner_id"`
	}

	err := db.Model(&Group{}).
		Select("id", "name", "owner_id").
		Order("id DESC").
		Find(&groups).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "检索群组失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
	})
}

// post /admin/forget_password
func forgetPassword(c *gin.Context) {
	var user struct {
		Id string `json:"id" binding:"required"`
	}

	a := jsonData(&user, c)
	if !a {
		return
	}

	db.Model(&User{}).
		Where("id = ?", user.Id).
		Updates(map[string]interface{}{
			"password": hashPassword(defaultPassword),
		})

	c.JSON(http.StatusOK, gin.H{
		"msg": "重置密码成功，新密码 111111",
	})
}

// post /admin/delete_user
func deleteUser(c *gin.Context) {
	var user struct {
		Id string `json:"id" binding:"required"`
	}

	a := jsonData(&user, c)
	if !a {
		return
	}

	db.Model(&User{}).Delete(&user)

	c.JSON(http.StatusOK, gin.H{
		"msg": "删除用户成功",
	})
}

// post /admin/delete_group
func deleteGroup(c *gin.Context) {
	var group struct {
		Id string `json:"id" binding:"required"`
	}

	a := jsonData(&group, c)
	if !a {
		return
	}

	db.Model(&Group{}).Delete(&group)

	c.JSON(http.StatusOK, gin.H{
		"msg": "删除群组成功",
	})
}

func jsonData[T any](model *T, c *gin.Context) bool {
	err := c.ShouldBindJSON(model)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效输入",
		})
		return false
	}
	return true
}

func timeoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if time.Now().After(timeout) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未经允许的访问"})
			return
		}

		c.Next()
	}
}
