package notification_module

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/arinsuda/movie-hub/internal/privacy_policy"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
)

type TokenVerifier interface {
	VerifyToken(token string) (uint, error)
}

type Hub struct {
	io          *socket.Server
	verifier    TokenVerifier
	mu          sync.RWMutex
	userSockets map[uint]map[string]bool
}

func NewHub(verifier TokenVerifier, allowedOrigins []string) *Hub {
	opts := socket.DefaultServerOptions()
	opts.SetCors(&types.Cors{
		Origin:      allowedOrigins,
		Credentials: true,
	})

	io := socket.NewServer(nil, opts)

	h := &Hub{
		io:          io,
		verifier:    verifier,
		userSockets: make(map[uint]map[string]bool),
	}

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)

		token := extractAccessTokenFromRequest(client.Handshake())
		if token == "" {
			_ = client.Emit("error", "unauthorized")
			client.Disconnect(true)
			return
		}

		userID, err := verifier.VerifyToken(token)
		if err != nil {
			_ = client.Emit("error", "unauthorized")
			client.Disconnect(true)
			return
		}

		socketID := string(client.Id())

		h.mu.Lock()
		if _, ok := h.userSockets[userID]; !ok {
			h.userSockets[userID] = make(map[string]bool)
		}
		h.userSockets[userID][socketID] = true
		h.mu.Unlock()

		client.On("disconnect", func(...any) {
			h.mu.Lock()
			defer h.mu.Unlock()
			if sockets, ok := h.userSockets[userID]; ok {
				delete(sockets, socketID)
				if len(sockets) == 0 {
					delete(h.userSockets, userID)
				}
			}
		})

		client.Join(roomFor(userID))
	})

	return h
}

func (h *Hub) UniqueOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userSockets)
}

func extractAccessTokenFromRequest(handshake *socket.Handshake) string {
	if handshake == nil {
		return ""
	}

	if handshake.Auth != nil {
		if tok, ok := handshake.Auth["token"].(string); ok && tok != "" {
			return tok
		}
	}

	if handshake.Query != nil {
		if tok, ok := handshake.Query["token"].(string); ok && tok != "" {
			return tok
		}
	}

	headers := handshake.Headers
	if headers == nil {
		return ""
	}

	if rawAuth, ok := headers["Authorization"]; ok {
		var authStr string
		switch v := rawAuth.(type) {
		case string:
			authStr = v
		case []string:
			if len(v) > 0 {
				authStr = v[0]
			}
		}
		if strings.HasPrefix(authStr, "Bearer ") {
			return strings.TrimPrefix(authStr, "Bearer ")
		}
	}
	if rawAuth, ok := headers["authorization"]; ok {
		var authStr string
		switch v := rawAuth.(type) {
		case string:
			authStr = v
		case []string:
			if len(v) > 0 {
				authStr = v[0]
			}
		}
		if strings.HasPrefix(authStr, "Bearer ") {
			return strings.TrimPrefix(authStr, "Bearer ")
		}
	}

	raw, ok := headers["Cookie"]
	if !ok {
		raw, ok = headers["cookie"]
	}
	if !ok {
		return ""
	}

	var cookieHeader string
	switch v := raw.(type) {
	case string:
		cookieHeader = v
	case []string:
		if len(v) == 0 {
			return ""
		}
		cookieHeader = v[0]
	default:
		return ""
	}

	if cookieHeader == "" {
		return ""
	}

	dummyReq := &http.Request{Header: http.Header{"Cookie": []string{cookieHeader}}}
	cookie, err := dummyReq.Cookie("access_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func roomFor(userID uint) socket.Room {
	return socket.Room("user:" + strconv.FormatUint(uint64(userID), 10))
}

func (h *Hub) Handler() http.Handler {
	return h.io.ServeHandler(nil)
}

func (h *Hub) EmitNew(userID uint, n NotificationResponse) {
	_ = h.io.To(roomFor(userID)).Emit("notification:new", n)
}

func (h *Hub) EmitRead(userID uint, ids []uint) {
	_ = h.io.To(roomFor(userID)).Emit("notification:read", map[string]any{"ids": ids})
}

func (h *Hub) EmitDeleted(userID uint, ids []uint) {
	_ = h.io.To(roomFor(userID)).Emit("notification:deleted", map[string]any{"ids": ids})
}

func (h *Hub) EmitUnreadCount(userID uint, count int64) {
	_ = h.io.To(roomFor(userID)).Emit("notification:unread_count", map[string]any{
		"unread_count": count,
	})
}

func (h *Hub) EmitFeedRefresh(userID uint) {
	_ = h.io.To(roomFor(userID)).Emit(privacy_policy.EventFeedRefresh, map[string]any{})
}

func (h *Hub) EmitFeedUpdated(userID uint, activityID uint, visibility string) {
	_ = h.io.To(roomFor(userID)).Emit(privacy_policy.EventFeedActivityUpdated, map[string]any{
		"id":         activityID,
		"visibility": visibility,
	})
}

func (h *Hub) EmitFeedDeleted(userID uint, activityID uint) {
	_ = h.io.To(roomFor(userID)).Emit(privacy_policy.EventFeedActivityRemoved, map[string]any{
		"id": activityID,
	})
}
