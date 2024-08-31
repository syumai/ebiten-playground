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

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/syumai/ebiten-playground/flappy/remote"
	"log"
	"net/http"
	"sync"
)

var errAlreadyClosed = errors.New("already closed")

type remoteServer struct {
	mu             sync.Mutex
	conn           *websocket.Conn
	connectionMode remote.ConnectionMode
}

func newRemoteServer(w http.ResponseWriter, r *http.Request) (*remoteServer, error) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to accept websocket connection: %w", err)
	}
	return &remoteServer{conn: c}, nil
}

func (r *remoteServer) Start(ctx context.Context) error {
	if err := r.readConnectionMessage(ctx); err != nil {
		return err
	}
	for {
		m, err := r.readMessage(ctx)
		if err != nil {
			return err
		}
		switch m.Type {
		case remote.MessageTypePublish:
		}
	}
}

func (r *remoteServer) startSubscription(ctx context.Context) error {

}

func (r *remoteServer) readConnectionMessage(ctx context.Context) error {
	var m remote.ConnectionMessage
	err := wsjson.Read(ctx, r.conn, &m)
	if err != nil {
		return fmt.Errorf("failed to read connection message: %w", err)
	}
	switch m.ConnectionMode {
	case remote.ConnectionModePublish:
		r.connectionMode = remote.ConnectionModePublish
	case remote.ConnectionModeSubscribe:
		r.connectionMode = remote.ConnectionModeSubscribe
	default:
		return fmt.Errorf("unknown connection mode %v", m.ConnectionMode)
	}
	return nil
}

func (r *remoteServer) readMessage(ctx context.Context) (*remote.Message, error) {
	var m remote.Message
	err := wsjson.Read(ctx, r.conn, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	switch m.Type {
	case remote.MessageTypePublish, remote.MessageTypeClose:
		// do nothing
	default:
		return nil, fmt.Errorf("unknown message type %v", m.Type)
	}
	return &m, nil
}

func (r *remoteServer) Close() error {
	if r.conn == nil {
		return errAlreadyClosed
	}
	c := r.conn
	r.conn = nil
	return c.CloseNow()
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	for {
		var v remote.Message
		err := wsjson.Read(r.Context(), c, &v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("failed to read websocket connection")
			return
		}
		if v.Type == remote.MessageTypeClose {
			break
		}
	}

	if err := c.Close(websocket.StatusNormalClosure, ""); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("failed to close websocket connection")
		return
	}
}
