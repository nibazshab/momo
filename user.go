package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// post /login
func userLogin(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	inputPassword := user.Password

	err = db.Select("password").First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, resp[any]{
			Code:    http.StatusUnauthorized,
			Message: "user invalid",
		})
		return
	}

	if !checkPassword(inputPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, resp[any]{
			Code:    http.StatusUnauthorized,
			Message: "password invalid",
		})
		return
	}

	token, err := generateToken(user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	c.JSON(http.StatusOK, resp[string]{
		Code: http.StatusOK,
		Data: token,
	})
}

// post /register
func userRegister(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	if !validateUserId(user.Id) {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "id invalid",
		})
		return
	}

	user.Name = strings.TrimSpace(user.Name)
	user.Password = hashPassword(user.Password)

	err = db.Create(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, resp[any]{
				Code:    http.StatusConflict,
				Message: "user already exists",
			})
		} else {
			c.JSON(http.StatusInternalServerError, resp[any]{
				Code:    http.StatusInternalServerError,
				Message: e,
			})
		}
		return
	}

	token, err := generateToken(user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	c.JSON(http.StatusOK, resp[string]{
		Code: http.StatusOK,
		Data: token,
	})
}

// get /api/v1/user/me
func userOwnInfo(c *gin.Context) {
	var user User
	user.Id = c.MustGet("userId").(int)

	err := getObjInfo(&user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, resp[any]{
				Code:    http.StatusNotFound,
				Message: "user invalid",
			})
		} else {
			c.JSON(http.StatusInternalServerError, resp[any]{
				Code:    http.StatusInternalServerError,
				Message: e,
			})
		}
		return
	}
	user.Password = ""

	c.JSON(http.StatusOK, resp[User]{
		Code: http.StatusOK,
		Data: user,
	})
}

// get /api/v1/user/list
func userList(c *gin.Context) {
	var users []User

	err := db.Model(&User{}).Select("id", "name").Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	c.JSON(http.StatusOK, resp[[]User]{
		Code: http.StatusOK,
		Data: users,
	})
}

// post /api/v1/user/rename
func resetName(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}
	user.Id = c.MustGet("userId").(int)

	rs := db.Model(user).Where("id = ?", user.Id).Update("name", strings.TrimSpace(user.Name))

	if rs.Error != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	if rs.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, resp[any]{
			Code:    http.StatusNotFound,
			Message: "user invalid",
		})
		return
	}

	c.JSON(http.StatusOK, resp[any]{
		Code:    http.StatusOK,
		Message: s,
	})
}

// post /api/v1/user/repassword
func resetPassword(c *gin.Context) {
	var password struct {
		Old string `json:"old_password"`
		New string `json:"new_password"`
	}
	err := c.ShouldBindJSON(&password)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	var user User
	user.Id = c.MustGet("userId").(int)

	err = db.Select("password").First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, resp[any]{
			Code:    http.StatusUnauthorized,
			Message: "user invalid",
		})
		return
	}

	if !checkPassword(password.Old, user.Password) {
		c.JSON(http.StatusUnauthorized, resp[any]{
			Code:    http.StatusUnauthorized,
			Message: "password invalid",
		})
		return
	}

	rs := db.Model(user).Where("id = ?", user.Id).Update("password", hashPassword(password.New))

	if rs.Error != nil || rs.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: s,
		})
		return
	}

	c.JSON(http.StatusOK, resp[any]{
		Code:    http.StatusOK,
		Message: s,
	})
}
