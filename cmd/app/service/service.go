package service

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gtihub.com/gin-websocket/cmd/app/protocol"
	"net/http"
	"sync/atomic"
	"time"
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

func (s *Service) storeChatMessage(m *protocol.Message) {
	data, err := proto.Marshal(m)
	NoError(err)
	NoError(s.r.LPush(s.c, s.redisChatHistoryKey(), string(data)).Err())
}

func (s *Service) getNextUid() uint32 {
	return atomic.AddUint32(&s.uid, 1)
}

func (s *Service) getAllHistoryMessages() []*protocol.Message {
	protoMessages, err := s.r.LRange(s.c, s.redisChatHistoryKey(), 0, -1).Result()
	NoError(err)

	var results = make([]*protocol.Message, len(protoMessages))

	for i, protoMessage := range protoMessages {
		var message protocol.Message
		NoError(proto.Unmarshal([]byte(protoMessage), &message))
		//NoError(json.Unmarshal([]byte(jsonMessage), &message))
		results[i] = &message
	}

	return results
}

func (s *Service) onNewMessage(m *protocol.Message) {
	if m == nil {
		return
	}
	s.broadcastMessage(m)
	s.storeChatMessage(m)
}

func (s *Service) broadcastMessage(m *protocol.Message) {
	clients := s.clients

	for _, client := range clients {
		go client.sendMessage(m)
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (s *Service) newClient(w http.ResponseWriter, r *http.Request) {
	wsupgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	wsupgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := wsupgrader.Upgrade(w, r, nil)
	NoError(err)

	cli := &client{
		conn:      conn,
		srv:       s,
		id:        s.getNextUid(),
		firstPing: true,
	}

	cli.conn.SetReadLimit(maxMessageSize)
	NoError(cli.conn.SetReadDeadline(time.Now().Add(pongWait)))
	cli.conn.SetPingHandler(func(string) error {
		NoError(cli.conn.SetReadDeadline(time.Now().Add(pongWait)))
		return nil
	})
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
