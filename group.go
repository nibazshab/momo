package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

// get /api/v1/group/member/:id
func groupMember(c *gin.Context) {
	stringGroupId := c.Param("id")
	groupId, err := strconv.Atoi(stringGroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id 必须是整数",
		})
		return
	}

	var users []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	err = db.Model(&GroupMember{}).
		Select("users.id, users.name").
		Joins("JOIN users ON group_members.user_id = users.id").
		Where("group_members.group_id = ?", groupId).
		Scan(&users).Error
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

// get /api/v1/group/info/:id
func groupInfo(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	stringGroupId := c.Param("id")
	groupId, err := strconv.Atoi(stringGroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id 必须是整数",
		})
		return
	}

	groupName, err := getNameById(&Group{}, groupId)
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
		Where("group_id = ? AND user_id = ?", groupId, userId).
		Count(&count).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查找群组失败",
		})
		return
	}
	isMember = count > 0

	c.JSON(http.StatusOK, gin.H{
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

// get /api/v1/group/join/:id
func memberJoin(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	stringGroupId := c.Param("id")
	groupId, err := strconv.Atoi(stringGroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id 必须是整数",
		})
		return
	}

	var count int64
	err = db.Model(&Group{}).
		Where("id = ?", groupId).
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
		Where("group_id = ? AND user_id = ?", groupId, userId).
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
		GroupId: groupId,
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

// get /api/v1/group/leave/:id
func memberLeave(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	stringGroupId := c.Param("id")
	groupId, err := strconv.Atoi(stringGroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id 必须是整数",
		})
		return
	}

	var group Group
	err = db.Select("owner_id").First(&group, "id = ?", groupId).Error
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
			err = tx.Delete(&Group{}, "id = ?", groupId).Error
			if err != nil {
				return err
			}

			err = tx.Where("conv_id = ?", groupId).Delete(&Msg{}).Error
			if err != nil {
				return err
			}
		} else {
			result := tx.Delete(&GroupMember{}, "group_id = ? AND user_id = ?", groupId, userId)

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

// get /api/v1/group/remove/:gid/:mid
func memberRemove(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	stringGroupId := c.Param("gid")
	stringMemberId := c.Param("mid")
	groupId, err := strconv.Atoi(stringGroupId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id 必须是整数",
		})
		return
	}
	memberId, err := strconv.Atoi(stringMemberId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id 必须是整数",
		})
		return
	}

	if memberId == userId {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不能移除自己",
		})
		return
	}

	var group Group
	err = db.Select("owner_id").First(&group, "id = ?", groupId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "群组不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "移除群员失败",
			})
		}
		return
	}

	if userId != group.OwnerId {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "权限不足",
		})
		return
	}

	err = db.Delete(&GroupMember{}, "group_id = ? AND user_id = ?", groupId, memberId).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "移除群员失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "移除群员成功",
	})
}
