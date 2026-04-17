package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
)

type ProgressEvent struct {
	DocumentID string `json:"document_id"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Message    string `json:"message"`
	Timestamp  int64  `json:"timestamp"`
}

func (h *Handler) DocumentProgressWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	documentID := r.PathValue("id")
	if documentID == "" {
		http.Error(w, "document id required", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionContextTakeover,
	})
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing")

	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		conn.Write(ctx, websocket.MessageText, mustMarshal(map[string]string{
			"error": "token required",
		}))
		conn.Close(websocket.StatusGoingAway, "token missing")
		return
	}

	claims, err := h.validateToken(tokenString)
	if err != nil {
		conn.Write(ctx, websocket.MessageText, mustMarshal(map[string]string{
			"error": "invalid token",
		}))
		conn.Close(websocket.StatusGoingAway, "invalid token")
		return
	}

	tenantID := claims.TenantID
	orgID := claims.OrgID

	if tenantID == "" || orgID == "" {
		conn.Write(ctx, websocket.MessageText, mustMarshal(map[string]string{
			"error": "invalid token claims",
		}))
		conn.Close(websocket.StatusGoingAway, "invalid claims")
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	currentStatus := ""

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, msg, prog := h.documentSvc.GetProcessingStatus(ctx, tenantID, orgID, documentID)

			if status != currentStatus {
				currentStatus = status

				event := ProgressEvent{
					DocumentID: documentID,
					Status:     status,
					Progress:   prog,
					Message:    msg,
					Timestamp:  time.Now().UnixMilli(),
				}

				if err := conn.Write(ctx, websocket.MessageText, mustMarshal(event)); err != nil {
					return
				}

				if status == "completed" || status == "failed" {
					return
				}
			}
		}
	}
}

type TokenClaims struct {
	UserID   string `json:"sub"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
	OrgID    string `json:"org_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (h *Handler) validateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.cfg.Auth.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	if claims.Issuer != h.cfg.Auth.JWTIssuer {
		return nil, jwt.ErrTokenInvalidIssuer
	}

	validAudience := false
	for _, aud := range claims.Audience {
		if aud == h.cfg.Auth.JWTAudience {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, jwt.ErrTokenInvalidAudience
	}

	return claims, nil
}

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
