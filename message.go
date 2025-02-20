package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsClientManager struct {
	clients map[int]map[*websocket.Conn]bool
	sync.RWMutex
}

var clientManager = WsClientManager{
	clients: make(map[int]map[*websocket.Conn]bool),
}

// ws /api/v1/ws/message
func messageHandler(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	stringConvId := c.Query("conv_id")
	if stringConvId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "缺少会话 ID",
		})
		return
	}

	convId, err := strconv.Atoi(stringConvId)
	if err != nil || convId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "会话 ID 无效",
		})
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "WebSocket 转换失败",
		})
		return
	}

	clientManager.Lock()
	if clientManager.clients[convId] == nil {
		clientManager.clients[convId] = make(map[*websocket.Conn]bool)
	}
	clientManager.clients[convId][ws] = true
	clientManager.Unlock()

	defer func() {
		clientManager.Lock()
		delete(clientManager.clients[convId], ws)
		if len(clientManager.clients[convId]) == 0 {
			delete(clientManager.clients, convId)
		}
		clientManager.Unlock()
		ws.Close()
	}()

	err = sendHistoricalMessages(ws, convId)
	if err != nil {
		log.Printf("Error sending history: %v", err)
	}

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		err = processMessage(message, userId, convId)
		if err != nil {
			log.Printf("Message processing error: %v", err)
		}
	}
}

func sendHistoricalMessages(ws *websocket.Conn, convId int) error {
	var messages []Msg
	err := db.Where("conv_id = ?", convId).
		Order("time ASC").
		Find(&messages).Error
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	for _, hisMsg := range messages {
		err = ws.WriteJSON(hisMsg)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}

	return nil
}

func processMessage(message []byte, userId, convId int) error {
	var msg struct {
		Text string `json:"text" binding:"required"`
		Type int    `json:"type"`
	}
	err := json.Unmarshal(message, &msg)
	if err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	if strings.TrimSpace(msg.Text) == "" {
		return fmt.Errorf("empty message content")
	}

	userName, _ := getNameById(&User{}, userId)

	newMsg := Msg{
		ConvId:   convId,
		UserId:   userId,
		UserName: userName,
		FmtTime:  time.Now().Format(time.DateTime),
		Text:     msg.Text,
		Type:     msg.Type,
	}

	err = db.Create(&newMsg).Error
	if err != nil {
		return fmt.Errorf("database save failed: %w", err)
	}

	clientManager.RLock()
	defer clientManager.RUnlock()

	for client := range clientManager.clients[convId] {
		go func(conn *websocket.Conn) {
			err = conn.WriteJSON(newMsg)
			if err != nil {
				log.Printf("Broadcast error: %v", err)
			}
		}(client)
	}

	return nil
}

// ws /api/v1/ws/convid
func convIdHandler(c *gin.Context) {
	userId := c.MustGet("userId").(int)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "WebSocket 转换失败",
		})
		return
	}

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var request struct {
			TargetId int `json:"target_id" binding:"required"`
		}
		err = json.Unmarshal(message, &request)
		if err != nil {
			ws.WriteJSON(gin.H{
				"error": "请求格式无效",
			})
			log.Printf("ConvId request error: %v", err)
			continue
		}

		if request.TargetId <= 0 {
			ws.WriteJSON(gin.H{
				"error": "目标用户 ID 无效",
			})
			log.Printf("ConvId request error: invalid target user")
			continue
		}

		var exists bool
		err = db.Model(&User{}).
			Select("count(*) > 0").
			Where("id = ?", request.TargetId).
			Find(&exists).Error
		if err != nil || !exists {
			ws.WriteJSON(gin.H{
				"error": "目标用户 ID 不存在",
			})
			log.Printf("ConvId request error: user not found")
			continue
		}

		convId := generateConvId(userId, request.TargetId)

		ws.WriteJSON(gin.H{
			"conv_id": convId,
		})
	}
}

func generateConvId(user1, user2 int) int {
	if user1 > user2 {
		user1, user2 = user2, user1
	}
	return (user1 << 32) | (user2 & 0xFFFFFFFF)
}
