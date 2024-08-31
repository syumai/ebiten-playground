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

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type MessageType string

const (
	MessageTypePublish MessageType = "message_type_publish"
	MessageTypeClose   MessageType = "message_type_close"
)

type Message struct {
	Type MessageType `json:"type"`
}

func writeMessage(ctx context.Context, conn *websocket.Conn, msg *Message) error {
	err := wsjson.Write(ctx, conn, msg)
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}
	return nil
}

func readMessage(ctx context.Context, conn *websocket.Conn) (*Message, error) {
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

type ConnectionMode string

const (
	ConnectionModeSubscribe ConnectionMode = "connection_mode_subscribe"
	ConnectionModePublish   ConnectionMode = "connection_mode_publish"
)

type ConnectionMessage struct {
	ConnectionMode ConnectionMode `json:"connectionMode"`
}

func writeConnectionMessage(ctx context.Context, conn *websocket.Conn, connectionMsg *ConnectionMessage) error {
	if err := wsjson.Write(ctx, conn, connectionMsg); err != nil {
		return fmt.Errorf("failed to write connection message: %w", err)
	}
	return nil
}

func readConnectionMessage(ctx context.Context, conn *websocket.Conn) (*ConnectionMessage, error) {
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

type AcceptationMessage struct {
	AcceptationResult string `json:"acceptationResult"`
}

const AcceptationResultOK = "OK"

func writeAcceptationMessage(ctx context.Context, conn *websocket.Conn) error {
	acceptationMsg := AcceptationMessage{AcceptationResult: "OK"}
	if err := wsjson.Write(ctx, conn, acceptationMsg); err != nil {
		return fmt.Errorf("failed to write acceptation message: %w", err)
	}
	return nil
}

func readAcceptationMessage(ctx context.Context, conn *websocket.Conn) error {
	var m AcceptationMessage
	if err := wsjson.Read(ctx, conn, &m); err != nil {
		return fmt.Errorf("failed to read acceptation message: %w", err)
	}
	if m.AcceptationResult != AcceptationResultOK {
		return fmt.Errorf("acceptation message is invalid: %s", m.AcceptationResult)
	}
	return nil
}
