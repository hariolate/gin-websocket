package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"sync/atomic"
)

type Service struct {
	id uuid.UUID

	r *redis.Client
	c context.Context

	*gin.Engine

	uid uint32

	clients map[uint32]*client
}

func FromConfig(c *Config, ctx context.Context) *Service {
	return &Service{
		id: uuid.New(),

		r: redis.NewClient(MustParseRedisURL(c.Storage.RedisURL)),
		c: ctx,

		Engine: gin.Default(),
		uid:    0,

		clients: make(map[uint32]*client),
	}
}

func (s *Service) redisChatHistoryKey() string {
	return fmt.Sprintf("chat:%s:history", s.id)
}

func (s *Service) storeChatMessage(m *Message) {
	NoError(s.r.LPush(s.c, s.redisChatHistoryKey(), m).Err())
}

func (s *Service) getNextUid() uint32 {
	return atomic.AddUint32(&s.uid, 1)
}

func (s *Service) getAllHistoryMessages() []*Message {
	jsonMessages, err := s.r.LRange(s.c, s.redisChatHistoryKey(), 0, -1).Result()
	NoError(err)

	var results = make([]*Message, len(jsonMessages))

	for i, jsonMessage := range jsonMessages {
		var message Message
		NoError(json.Unmarshal([]byte(jsonMessage), &message))
		results[i] = &message
	}

	return results
}

func (s *Service) onNewMessage(m *Message) {
	if m == nil {
		return
	}
	s.broadcastMessage(m)
	s.storeChatMessage(m)
}

func (s *Service) broadcastMessage(m *Message) {
	clients := s.clients

	for _, client := range clients {
		go client.sendMessage(m)
	}
}

func (s *Service) newClient(w http.ResponseWriter, r *http.Request) {
	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := wsupgrader.Upgrade(w, r, nil)
	NoError(err)

	cli := &client{
		conn:      conn,
		srv:       s,
		id:        s.getNextUid(),
		firstPing: true,
	}

	cli.setupWorkers()
	s.clients[cli.id] = cli
}

func (s *Service) removeClient(c *client) {
	delete(s.clients, c.id)
}

func (s *Service) Handler(c *gin.Context) {
	s.newClient(c.Writer, c.Request)
	c.Status(http.StatusOK)
}
