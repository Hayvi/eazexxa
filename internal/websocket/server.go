package websocket

import (
	"net/http"
	"strings"

	"github.com/betpro/server/internal/services"
	"github.com/betpro/server/pkg/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	hub         *Hub
	authService *services.AuthService
}

func NewServer(hub *Hub, authService *services.AuthService) *Server {
	return &Server{
		hub:         hub,
		authService: authService,
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		logger.Warn("websocket connection rejected: no token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := s.authService.VerifyToken(token)
	if err != nil {
		logger.Warn("websocket connection rejected: invalid token", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("failed to upgrade connection", "error", err)
		return
	}

	client := &Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		hub:    s.hub,
	}

	s.hub.register <- client

	go client.WritePump()
	go client.ReadPump()

	logger.Info("websocket connection established", "user_id", claims.UserID)
}
