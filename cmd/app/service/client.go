package service

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/gorilla/websocket"
	"gtihub.com/gin-websocket/cmd/app/protocol"
	"log"
	"time"
)

type client struct {
	conn *websocket.Conn
	srv  *Service

	id        uint32
	firstPing bool
}

func (c *client) redisTimeoutKey() string {
	return fmt.Sprintf("client:%d:timeout", c.id)
}

//const clientTimeout = time.Minute * 2

func (c *client) setupWorkers() {
	//go c.timeoutWorker()
	go c.pingWorker()
	go c.receiveWorker()
}

//func (c *client) timeoutWorker() {
//	NoError(c.srv.r.Set(c.srv.c, c.redisTimeoutKey(), 1, clientTimeout).Err())
//
//	for {
//		if c.srv.r.Get(c.srv.c, c.redisTimeoutKey()).Err() != nil {
//			_ = c.conn.Close()
//			break
//		}
//	}
//}

func (c *client) receiveWorker() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("client %d crashed: %s\n", c.id, r)
		}
		c.srv.removeClient(c)
		_ = c.conn.Close()
	}()

	for {
		t, msg, err := c.conn.ReadMessage()
		NoError(err)
		go c.srv.onNewMessage(c.handleNewMessage(t, msg))
	}
}

func (c *client) handleNewMessage(messageType int, data []byte) *protocol.Message {
	//if messageType == websocket.PingMessage {
	//	c.handlePing()
	//	return nil
	//}

	//var raw RawMessage
	//NoError(json.Unmarshal(data, &raw))

	var raw protocol.RawMessage
	NoError(proto.Unmarshal(data, &raw))
	//if messageType != websocket.TextMessage {
	//	raw.Message = "--unsupported message--"
	//}

	//return &Message{
	//	UID:       c.id,
	//	Raw:       raw,
	//	Timestamp: time.Now(),
	//}

	now, err := ptypes.TimestampProto(time.Now())
	NoError(err)

	return &protocol.Message{
		Uid:       c.id,
		Raw:       nil,
		Timestamp: now,
	}
}

//func (c *client) handlePing() {
//	writer, err := c.conn.NextWriter(websocket.PongMessage)
//	NoError(err)
//	_, err = writer.Write([]byte("pong"))
//	NoError(err)
//
//	NoError(c.srv.r.Set(c.srv.c, c.redisTimeoutKey(), 1, clientTimeout).Err())
//
//	if c.firstPing {
//		go c.sendMessages(c.srv.getAllHistoryMessages())
//		c.firstPing = false
//	}
//}

func (c *client) pingWorker() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		<-ticker.C
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
		ticker.Reset(pingPeriod)
	}
}

func (c *client) sendMessage(m *protocol.Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("client %d crashed: %s\n", c.id, r)
			c.srv.removeClient(c)
			_ = c.conn.Close()
		}
	}()
	data, err := proto.Marshal(m)
	NoError(err)
	NoError(c.conn.SetWriteDeadline(time.Now().Add(writeWait)))
	NoError(c.conn.WriteMessage(websocket.BinaryMessage, data))
}

func (c *client) sendMessages(ms []*protocol.Message) {
	for _, m := range ms {
		go c.sendMessage(m)
	}
}
