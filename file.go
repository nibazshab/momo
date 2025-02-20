package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxUploadSize = 100 << 20 // 100mb
	uploadPath    = "attachments"
)

func initDir() {
	err := os.MkdirAll("attachments", 0o755)
	if err != nil {
		panic("[err]" + err.Error())
	}
}

// post /api/v1/upload
func upFileHandler(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	err := c.Request.ParseMultipartForm(maxUploadSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件最大 100mb",
		})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件表单无效",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文件打开失败",
		})
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文件验证失败",
		})
		return
	}

	mimeType := http.DetectContentType(buffer)
	var fileType int
	if strings.HasPrefix(mimeType, "image/") {
		fileType = 1
	} else {
		fileType = 2
	}

	safeName := uuid.New().String()
	filePath := filepath.Join(uploadPath, safeName)

	err = c.SaveUploadedFile(fileHeader, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文件保存失败",
		})
		return
	}

	fileSize := fileHeader.Size >> 10

	fileRecord := File{
		UUID:         safeName,
		OriginalName: fileHeader.Filename,
		Type:         fileType,
		Size:         fileSize,
		UserId:       userId,
	}

	err = db.Create(&fileRecord).Error
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "记录文件元数据失败",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"uuid": fmt.Sprintf("%s.%dkb.%s", safeName, fileSize, fileHeader.Filename),
		"type": fileType,
	})
}

// get /api/v1/files/:filename
func downFileHandler(c *gin.Context) {
	fileUuid := c.Param("filename")
	if fileUuid == "" || strings.Contains(fileUuid, "..") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件名无效",
		})
		return
	}

	var fileRecord File
	err := db.Where("uuid = ?", fileUuid).First(&fileRecord).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文件不存在",
		})
		return
	}

	filePath := filepath.Join(uploadPath, fileUuid)
	c.FileAttachment(filePath, fileRecord.OriginalName)
}
