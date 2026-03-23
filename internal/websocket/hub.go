package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/betpro/server/internal/services"
	"github.com/betpro/server/pkg/logger"
	"github.com/gorilla/websocket"
)

type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	hub    *Hub
}

type Hub struct {
	clients    map[string]map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	redis      *services.RedisClient
	channel    string
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	UserID  string      `json:"-"`
}

func NewHub(redis *services.RedisClient, channel string) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      redis,
		channel:    channel,
	}
}

func (h *Hub) Run(ctx context.Context) {
	if h.redis.IsEnabled() {
		go h.subscribeRedis(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			logger.Debug("client registered", "user_id", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			close(client.Send)
			h.mu.Unlock()
			logger.Debug("client unregistered", "user_id", client.UserID)

		case message := <-h.broadcast:
			h.sendMessage(message)
		}
	}
}

func (h *Hub) sendMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("failed to marshal message", "error", err)
		return
	}

	if msg.UserID != "" {
		h.sendToUser(msg.UserID, data)
	} else {
		h.sendToAll(data)
	}

	if h.redis.IsEnabled() {
		_ = h.redis.Publish(context.Background(), h.channel, data)
	}
}

func (h *Hub) sendToUser(userID string, data []byte) {
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			h.mu.Lock()
			delete(h.clients[userID], client)
			h.mu.Unlock()
		}
	}
}

func (h *Hub) sendToAll(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
			}
		}
	}
}

func (h *Hub) subscribeRedis(ctx context.Context) {
	pubsub := h.redis.Subscribe(ctx, h.channel)
	if pubsub == nil {
		return
	}
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var message Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				logger.Error("failed to unmarshal redis message", "error", err)
				continue
			}

			if message.UserID != "" {
				h.sendToUser(message.UserID, []byte(msg.Payload))
			} else {
				h.sendToAll([]byte(msg.Payload))
			}
		}
	}
}

func (h *Hub) Broadcast(msgType string, payload interface{}) {
	h.broadcast <- &Message{
		Type:    msgType,
		Payload: payload,
	}
}

func (h *Hub) BroadcastToUser(userID, msgType string, payload interface{}) {
	h.broadcast <- &Message{
		Type:    msgType,
		Payload: payload,
		UserID:  userID,
	}
}

func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	totalConnections := 0
	for _, clients := range h.clients {
		totalConnections += len(clients)
	}

	return map[string]interface{}{
		"total_users":       len(h.clients),
		"total_connections": totalConnections,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket error", "error", err)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
