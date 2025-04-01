package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// get /api/v1/group/list
func groupListAsMember(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	var groups []Group

	err := db.Model(&GroupMember{}).
		Select("groups.id, groups.name, groups.owner_id").
		Joins("JOIN groups ON groups.id = group_members.group_id").
		Where("group_members.user_id = ?", userId).
		Find(&groups).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	c.JSON(http.StatusOK, resp[[]Group]{
		Code: http.StatusOK,
		Data: groups,
	})
}

// post /api/v1/group/create
func groupCreate(c *gin.Context) {
	var group Group
	err := c.ShouldBindJSON(&group)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	if group.Id < 100000 || group.Id > 999999 {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "id invalid",
		})
		return
	}

	group.OwnerId = c.MustGet("userId").(int)
	group.Name = strings.TrimSpace(group.Name)

	err = db.Create(group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, resp[any]{
				Code:    http.StatusConflict,
				Message: "group already exists",
			})
		} else {
			c.JSON(http.StatusInternalServerError, resp[any]{
				Code:    http.StatusInternalServerError,
				Message: e,
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp[any]{
		Code:    http.StatusOK,
		Message: s,
	})
}

func (g *Group) AfterCreate(tx *gorm.DB) (err error) {
	groupMember := GroupMember{
		GroupId: g.Id,
		UserId:  g.OwnerId,
	}

	err = tx.Create(&groupMember).Error
	if err != nil {
		return err
	}
	return nil
}

// post /api/v1/group/info
func groupInfo(c *gin.Context) {
	var group Group
	err := c.ShouldBindJSON(&group)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	err = getObjInfo(&group)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, resp[any]{
				Code:    http.StatusNotFound,
				Message: "group invalid",
			})
		} else {
			c.JSON(http.StatusInternalServerError, resp[any]{
				Code:    http.StatusInternalServerError,
				Message: e,
			})
		}
		return
	}

	userId := c.MustGet("userId").(int)

	var isMember string
	err = validateGroupMember(group.Id, userId)
	if err != nil {
		isMember = "n"
	} else {
		isMember = "y"
	}

	c.JSON(http.StatusOK, resp[Group]{
		Code:    http.StatusOK,
		Message: isMember,
		Data:    group,
	})
}

// post /api/v1/group/join
func groupJoin(c *gin.Context) {
	var groupMember GroupMember
	err := c.ShouldBindJSON(&groupMember)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	err = validateGroup(groupMember.GroupId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	groupMember.UserId = c.MustGet("userId").(int)

	err = validateGroupMember(groupMember.GroupId, groupMember.UserId)
	if err == nil {
		c.JSON(http.StatusConflict, resp[any]{
			Code:    http.StatusConflict,
			Message: "already joined",
		})
		return
	}

	err = db.Create(groupMember).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	c.JSON(http.StatusOK, resp[any]{
		Code:    http.StatusOK,
		Message: s,
	})
}

// post /api/v1/group/member
func memberList(c *gin.Context) {
	var groupMember GroupMember
	err := c.ShouldBindJSON(&groupMember)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}
	groupMember.UserId = c.MustGet("userId").(int)

	err = validateGroupMember(groupMember.GroupId, groupMember.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	var users []User

	err = db.Model(groupMember).
		Select("users.id, users.name").
		Joins("JOIN users ON users.id = group_members.user_id").
		Where("group_members.group_id = ?", groupMember.GroupId).
		Scan(&users).Error
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

// post /api/v1/group/leave
func memberLeave(c *gin.Context) {
	var groupMember GroupMember
	err := c.ShouldBindJSON(&groupMember)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	err = validateGroupMember(groupMember.GroupId, groupMember.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	groupMember.UserId = c.MustGet("userId").(int)

	group := Group{
		Id: groupMember.GroupId,
	}
	err = getObjInfo(&group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if groupMember.UserId == group.OwnerId {
			err = tx.Delete(group).Error
			if err != nil {
				return err
			}
			err = tx.Where("conv_id = ?", group.Id).Delete(&Msg{}).Error
			if err != nil {
				return err
			}
		} else {
			err = tx.Delete(groupMember).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	c.JSON(http.StatusOK, resp[any]{
		Code:    http.StatusOK,
		Message: s,
	})
}

// post /api/v1/group/remove
func memberRemove(c *gin.Context) {
	var groupMember GroupMember
	err := c.ShouldBindJSON(&groupMember)
	if err != nil {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "input invalid",
		})
		return
	}

	group := Group{
		Id: groupMember.GroupId,
	}
	err = getObjInfo(&group)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, resp[any]{
				Code:    http.StatusConflict,
				Message: "group invalid",
			})
		} else {
			c.JSON(http.StatusInternalServerError, resp[any]{
				Code:    http.StatusInternalServerError,
				Message: e,
			})
		}
		return
	}

	err = validateGroupMember(groupMember.GroupId, groupMember.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	userId := c.MustGet("userId").(int)

	if userId != group.OwnerId {
		c.JSON(http.StatusForbidden, resp[any]{
			Code:    http.StatusForbidden,
			Message: "permission denied",
		})
		return
	}

	if userId == groupMember.UserId {
		c.JSON(http.StatusBadRequest, resp[any]{
			Code:    http.StatusBadRequest,
			Message: "yourself invalid",
		})
		return
	}

	err = db.Delete(groupMember).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, resp[any]{
			Code:    http.StatusInternalServerError,
			Message: e,
		})
		return
	}

	c.JSON(http.StatusOK, resp[any]{
		Code:    http.StatusOK,
		Message: s,
	})
}
