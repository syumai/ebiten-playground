// Copyright 2024 syumai
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

type Server struct {
	mu            sync.RWMutex
	subscriptions map[string]chan<- *Message
}

func NewServer() *Server {
	return &Server{subscriptions: map[string]chan<- *Message{}}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := s.accept(w, r); err != nil {
		fmt.Fprintf(w, "failed to accept websocket connection: %v", err)
	}
}

func (s *Server) accept(w http.ResponseWriter, r *http.Request) error {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return fmt.Errorf("failed to accept websocket connection: %w", err)
	}
	defer conn.CloseNow()
	return s.start(r.Context(), conn)
}

func (s *Server) start(ctx context.Context, conn *websocket.Conn) error {
	connectionMsg, err := s.readConnectionMessage(ctx, conn)
	if err != nil {
		return err
	}

	switch connectionMsg.ConnectionMode {
	case ConnectionModePublish, ConnectionModeSubscribe:
		if err := s.writeAcceptationMessage(ctx, conn); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown connection mode %v", connectionMsg.ConnectionMode)
	}

	switch connectionMsg.ConnectionMode {
	case ConnectionModePublish:
		return s.startPublishing(ctx, conn)
	case ConnectionModeSubscribe:
		return s.startSubscription(ctx, conn)
	}
	return nil // unreachable
}

func (s *Server) readConnectionMessage(ctx context.Context, conn *websocket.Conn) (*ConnectionMessage, error) {
	var connectionMsg ConnectionMessage
	if err := wsjson.Read(ctx, conn, &connectionMsg); err != nil {
		return nil, fmt.Errorf("failed to read connection message: %w", err)
	}
	switch connectionMsg.ConnectionMode {
	case ConnectionModePublish, ConnectionModeSubscribe:
		// do nothing
	default:
		return nil, fmt.Errorf("unknown connection mode %v", connectionMsg.ConnectionMode)
	}
	return &connectionMsg, nil
}

func (s *Server) writeAcceptationMessage(ctx context.Context, conn *websocket.Conn) error {
	acceptationMsg := AcceptationMessage{AcceptationResult: "OK"}
	if err := wsjson.Write(ctx, conn, acceptationMsg); err != nil {
		return fmt.Errorf("failed to write acceptation message: %w", err)
	}
	return nil
}

func (s *Server) startPublishing(ctx context.Context, conn *websocket.Conn) error {
LOOP:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		m, err := s.readMessage(ctx, conn)
		if err != nil {
			return err
		}
		switch m.Type {
		case MessageTypePublish:
			s.mu.RLock()
			for _, ch := range s.subscriptions {
				// TODO: make parallel
				ch <- m
			}
			s.mu.RUnlock()
		case MessageTypeClose:
			break LOOP
		default:
			return fmt.Errorf("unexpected message type %v", m.Type)
		}
	}
	return nil
}

func (s *Server) startSubscription(ctx context.Context, conn *websocket.Conn) error {
	clientID := uuid.NewString()

	var chMu sync.Mutex
	subscriberCh := make(chan struct {
		msg *Message
		err error
	})
	publisherCh := make(chan *Message, 10)

	s.mu.Lock()
	s.subscriptions[clientID] = publisherCh
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.subscriptions, clientID)
		s.mu.Unlock()
		chMu.Lock()
		defer chMu.Unlock()
		close(publisherCh)
		close(subscriberCh)
		publisherCh = nil
		subscriberCh = nil
	}()

	go func() {
		msg, err := s.readMessage(ctx, conn)
		chMu.Lock()
		defer chMu.Unlock()
		select {
		// subscriberCh becomes nil and blocks forever after closing channel, so default case will proceed
		case subscriberCh <- struct {
			msg *Message
			err error
		}{
			msg: msg,
			err: err,
		}:
		default:
		}
	}()

LOOP:
	for {
		select {
		case msg := <-publisherCh:
			switch msg.Type {
			case MessageTypePublish:
				if err := wsjson.Write(ctx, conn, msg); err != nil {
					return fmt.Errorf("failed to write message: %w", err)
				}
			default:
				return fmt.Errorf("unexpected publisher message type %v", msg.Type)
			}
		case result := <-subscriberCh:
			if result.err != nil {
				return fmt.Errorf("unexpected subscriber error: %w", result.err)
			}
			switch result.msg.Type {
			case MessageTypeClose:
				break LOOP
			default:
				return fmt.Errorf("unexpected subscriber message type %v", result.msg.Type)
			}
		case <-time.After(30 * time.Minute):
			return fmt.Errorf("subscription timed out")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *Server) readMessage(ctx context.Context, conn *websocket.Conn) (*Message, error) {
	var m Message
	err := wsjson.Read(ctx, conn, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	switch m.Type {
	case MessageTypePublish, MessageTypeClose:
		// do nothing
	default:
		return nil, fmt.Errorf("unknown message type %v", m.Type)
	}
	return &m, nil
}
