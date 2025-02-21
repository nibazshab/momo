package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// get /api/v1/group/lists
func groupListAsMember(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var groups []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		OwnerId int    `json:"owner_id"`
	}

	err := db.Model(&GroupMember{}).
		Select("groups.id, groups.name, groups.owner_id").
		Joins("JOIN groups ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userId).
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

// post /api/v1/group/info
func groupInfo(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var request struct {
		ID int `json:"id" binding:"required"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式无效",
		})
		return
	}

	groupName, err := getNameById(&Group{}, request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "群组不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "查找群组失败",
			})
		}
		return
	}

	var isMember bool
	var count int64
	err = db.Model(&GroupMember{}).
		Where("group_id = ? AND user_id = ?", request.ID, userId).
		Count(&count).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查找群组失败",
		})
		return
	}
	isMember = count > 0

	c.JSON(http.StatusOK, gin.H{
		"id":        request.ID,
		"name":      groupName,
		"is_member": isMember,
	})
}

// post /api/v1/group/create
func groupCreate(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var request struct {
		ID   int    `json:"id" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式无效",
		})
		return
	}

	if request.ID < 100000 || request.ID > 999999 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "群组 ID 必须是 6 位数字",
		})
		return
	}

	group := Group{
		ID:      request.ID,
		Name:    request.Name,
		OwnerId: userId,
	}

	err = db.Create(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "群组已存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "创建群组失败",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg": "创建群组成功",
	})
}

func (g *Group) AfterCreate(tx *gorm.DB) (err error) {
	groupMember := GroupMember{
		GroupId: g.ID,
		UserId:  g.OwnerId,
	}

	err = tx.Create(&groupMember).Error
	if err != nil {
		return err
	}
	return nil
}

// post /api/v1/group/join
func memberJoin(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var request struct {
		ID int `json:"id" binding:"required"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式无效",
		})
		return
	}

	var count int64
	err = db.Model(&Group{}).
		Where("id = ?", request.ID).
		Count(&count).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "加入群组失败",
		})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "群组不存在",
		})
		return
	}

	var existing int64
	err = db.Model(&GroupMember{}).
		Where("group_id = ? AND user_id = ?", request.ID, userId).
		Count(&existing).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "加入群组失败",
		})
		return
	}

	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "已是群组成员",
		})
		return
	}

	member := GroupMember{
		GroupId: request.ID,
		UserId:  userId,
	}

	err = db.Create(&member).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "加入群组失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg": "加入群组成功",
	})
}

// post /api/v1/group/leave
func memberLeave(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var request struct {
		ID int `json:"id" binding:"required"`
	}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求格式无效",
		})
		return
	}

	var group Group
	err = db.Select("owner_id").First(&group, "id = ?", request.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "群组不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "退出群组失败",
			})
		}
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if group.OwnerId == userId {
			err = tx.Delete(&Group{}, "id = ?", request.ID).Error
			if err != nil {
				return err
			}

			err = tx.Where("conv_id = ?", request.ID).Delete(&Msg{}).Error
			if err != nil {
				return err
			}
		} else {
			result := tx.Delete(&GroupMember{}, "group_id = ? AND user_id = ?", request.ID, userId)

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				return fmt.Errorf("0")
			}
		}
		return nil
	})
	if err != nil {
		if err.Error() == "0" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "不是群组成员",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "退出群组失败",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "退出群组成功",
	})
}
