package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://livepoll-frontend-st0u.onrender.com"
	},
}

func (h *PollHandler) Realtime(c *gin.Context) {
	pollID := c.Param("id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	channel := "poll:" + pollID + ":updates"

	pubsub := h.RedisClient.Subscribe(context.Background(), channel)
	defer pubsub.Close()

	for {
		message, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			return
		}

		err = conn.WriteMessage(
			websocket.TextMessage,
			[]byte(message.Payload),
		)

		if err != nil {
			return
		}
	}
}