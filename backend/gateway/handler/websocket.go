package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/gorilla/websocket"
)

type WsClient struct {
	UserID   int64
	Conn     *websocket.Conn
	Send     chan []byte
	RoomIDs  map[string]bool
	LastPing time.Time
}

type WsMessage struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"roomId,omitempty"`
	ToUser  int64       `json:"toUser,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clients     = make(map[int64]*WsClient)
	clientsLock sync.RWMutex

	rooms     = make(map[string]map[int64]bool)
	roomsLock sync.RWMutex
)

func HandleWebSocket(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.UserID == 0 {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("未授权"))
		return
	}

	conn, err := upgrader.Upgrade(c.GetWriter(), c.GetRequest(), nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", logger.Error(err))
		return
	}

	client := &WsClient{
		UserID:   userInfo.UserID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		RoomIDs:  make(map[string]bool),
		LastPing: time.Now(),
	}

	clientsLock.Lock()
	clients[userInfo.UserID] = client
	clientsLock.Unlock()

	logger.Info("WebSocket connected", logger.Int64("userId", userInfo.UserID))

	go client.readPump()
	go client.writePump()

	sendWelcomeMessage(client)

	BroadcastOnlineCount()
}

func HandleCaseWebSocket(ctx context.Context, c *app.RequestContext) {
	caseID := c.Param("id")
	HandleWebSocket(ctx, c)

	if client, ok := clients[middleware.GetUserInfo(c).UserID]; ok {
		JoinRoom(client, fmt.Sprintf("case:%s", caseID))
	}
}

func HandleVideoWebSocket(ctx context.Context, c *app.RequestContext) {
	roomID := c.Param("roomId")
	HandleWebSocket(ctx, c)

	if client, ok := clients[middleware.GetUserInfo(c).UserID]; ok {
		JoinRoom(client, fmt.Sprintf("video:%s", roomID))
	}
}

func (client *WsClient) readPump() {
	defer func() {
		client.cleanup()
	}()

	client.Conn.SetReadLimit(4096)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.LastPing = time.Now()
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msgBytes, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket read error", logger.Error(err))
			}
			break
		}

		var msg WsMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			logger.Error("WebSocket message parse error", logger.Error(err))
			continue
		}

		client.handleMessage(&msg)
	}
}

func (client *WsClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.cleanup()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("WebSocket write error", logger.Error(err))
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

			if time.Since(client.LastPing) > 120*time.Second {
				logger.Info("WebSocket timeout, closing connection", 
					logger.Int64("userId", client.UserID))
				return
			}
		}
	}
}

func (client *WsClient) handleMessage(msg *WsMessage) {
	switch msg.Type {
	case "ping":
		client.LastPing = time.Now()
		sendMessage(client, &WsMessage{Type: "pong", Data: time.Now().Unix()})

	case "join_room":
		if msg.RoomID != "" {
			JoinRoom(client, msg.RoomID)
		}

	case "leave_room":
		if msg.RoomID != "" {
			LeaveRoom(client, msg.RoomID)
		}

	case "case_message":
		if msg.RoomID != "" && msg.Message != "" {
			BroadcastToRoom(msg.RoomID, &WsMessage{
				Type:    "case_message",
				RoomID:  msg.RoomID,
				ToUser:  msg.ToUser,
				Data:    msg.Data,
				Message: msg.Message,
			})
		}

	case "video_offer", "video_answer", "video_candidate":
		if msg.RoomID != "" {
			BroadcastToRoom(msg.RoomID, &WsMessage{
				Type:    msg.Type,
				RoomID:  msg.RoomID,
				Data:    msg.Data,
			})
		}

	case "subscribe":
		cacheKey := fmt.Sprintf("ws:subscribe:%d", client.UserID)
		cache.Set(context.Background(), cacheKey, "1", 24*time.Hour)

	default:
		logger.Warn("Unknown WebSocket message type", 
			logger.String("type", msg.Type),
			logger.Int64("userId", client.UserID))
	}
}

func (client *WsClient) cleanup() {
	clientsLock.Lock()
	if _, ok := clients[client.UserID]; ok {
		delete(clients, client.UserID)
		close(client.Send)

		for roomID := range client.RoomIDs {
			roomsLock.Lock()
			if room, ok := rooms[roomID]; ok {
				delete(room, client.UserID)
				if len(room) == 0 {
					delete(rooms, roomID)
				}
			}
			roomsLock.Unlock()
		}
	}
	clientsLock.Unlock()

	client.Conn.Close()
	logger.Info("WebSocket disconnected", logger.Int64("userId", client.UserID))

	BroadcastOnlineCount()
}

func JoinRoom(client *WsClient, roomID string) {
	roomsLock.Lock()
	defer roomsLock.Unlock()

	if _, ok := rooms[roomID]; !ok {
		rooms[roomID] = make(map[int64]bool)
	}
	rooms[roomID][client.UserID] = true
	client.RoomIDs[roomID] = true

	logger.Info("User joined room", 
		logger.Int64("userId", client.UserID),
		logger.String("roomId", roomID))

	sendMessage(client, &WsMessage{
		Type:    "joined_room",
		RoomID:  roomID,
		Message: "已加入房间",
	})
}

func LeaveRoom(client *WsClient, roomID string) {
	roomsLock.Lock()
	defer roomsLock.Unlock()

	if room, ok := rooms[roomID]; ok {
		delete(room, client.UserID)
		if len(room) == 0 {
			delete(rooms, roomID)
		}
	}
	delete(client.RoomIDs, roomID)

	logger.Info("User left room", 
		logger.Int64("userId", client.UserID),
		logger.String("roomId", roomID))
}

func sendMessage(client *WsClient, msg *WsMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Marshal message failed", logger.Error(err))
		return
	}

	select {
	case client.Send <- msgBytes:
	default:
		logger.Warn("Send channel full, dropping message", 
			logger.Int64("userId", client.UserID))
	}
}

func SendToUser(userID int64, msg *WsMessage) {
	clientsLock.RLock()
	defer clientsLock.RUnlock()

	if client, ok := clients[userID]; ok {
		sendMessage(client, msg)
	}
}

func Broadcast(msg *WsMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Marshal message failed", logger.Error(err))
		return
	}

	clientsLock.RLock()
	defer clientsLock.RUnlock()

	for _, client := range clients {
		select {
		case client.Send <- msgBytes:
		default:
			logger.Warn("Send channel full, dropping broadcast", 
				logger.Int64("userId", client.UserID))
		}
	}
}

func BroadcastToRoom(roomID string, msg *WsMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Error("Marshal message failed", logger.Error(err))
		return
	}

	roomsLock.RLock()
	room, ok := rooms[roomID]
	roomsLock.RUnlock()

	if !ok {
		return
	}

	clientsLock.RLock()
	defer clientsLock.RUnlock()

	for userID := range room {
		if client, ok := clients[userID]; ok {
			select {
			case client.Send <- msgBytes:
			default:
				logger.Warn("Send channel full, dropping room message", 
					logger.Int64("userId", userID))
			}
		}
	}
}

func BroadcastOnlineCount() {
	clientsLock.RLock()
	count := len(clients)
	clientsLock.RUnlock()

	Broadcast(&WsMessage{
		Type: "online_count",
		Data: count,
	})
}

func sendWelcomeMessage(client *WsClient) {
	sendMessage(client, &WsMessage{
		Type:    "welcome",
		Message: "连接成功",
		Data: map[string]interface{}{
			"userId":    client.UserID,
			"timestamp": time.Now().Unix(),
		},
	})
}

func GetOnlineUsers(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限查看在线用户"))
		return
	}

	clientsLock.RLock()
	defer clientsLock.RUnlock()

	var userIDs []int64
	for userID := range clients {
		userIDs = append(userIDs, userID)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"count":   len(clients),
		"userIds": userIDs,
	}))
}

func BroadcastNotification(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限发送广播"))
		return
	}

	var req struct {
		Title   string      `json:"title" binding:"required"`
		Content string      `json:"content" binding:"required"`
		Type    string      `json:"type"`
		Data    interface{} `json:"data"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	msg := &WsMessage{
		Type:    "notification",
		Message: req.Title,
		Data: map[string]interface{}{
			"title":   req.Title,
			"content": req.Content,
			"type":    req.Type,
			"data":    req.Data,
			"sentBy":  userInfo.RealName,
			"sentAt":  time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	Broadcast(msg)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "广播发送成功"))
}
