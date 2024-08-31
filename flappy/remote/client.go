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
	"errors"
	"fmt"
	"iter"

	"github.com/coder/websocket"
)

const WSServerPort = "7382"

var ErrAlreadyClosed = errors.New("already closed")

type Client struct {
	conn *websocket.Conn
}

func NewClient(ctx context.Context) (*Client, error) {
	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://localhost:%s", WSServerPort), nil)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Subscribe(ctx context.Context) (iter.Seq2[*Message, error], error) {
	connectionMsg := ConnectionMessage{ConnectionMode: ConnectionModeSubscribe}
	if err := writeConnectionMessage(ctx, c.conn, &connectionMsg); err != nil {
		return nil, err
	}
	return func(yield func(*Message, error) bool) {
		for {
			msg, err := func() (*Message, error) {
				select {
				case <-ctx.Done():
					if err := writeMessage(ctx, c.conn, &Message{Type: MessageTypeClose}); err != nil {
						return nil, err
					}
					return nil, ctx.Err()
				default:
				}
				msg, err := readMessage(ctx, c.conn)
				if err != nil {
					return nil, err
				}
				if msg.Type != MessageTypePublish {
					return nil, fmt.Errorf("unexpected message type %v", msg.Type)
				}
				return msg, nil
			}()

			wantNext := yield(msg, err)
			if !wantNext {
				return
			}
			if err != nil {
				return
			}
		}
	}, nil
}

type Publisher struct {
	conn *websocket.Conn
}

func (c *Client) NewPublisher(ctx context.Context) (*Publisher, error) {
	connectionMsg := ConnectionMessage{ConnectionMode: ConnectionModePublish}
	if err := writeConnectionMessage(ctx, c.conn, &connectionMsg); err != nil {
		return nil, err
	}
	return &Publisher{conn: c.conn}, nil
}

func (pub *Publisher) Publish(ctx context.Context, msg *Message) error {
	return writeMessage(ctx, pub.conn, msg)
}

func (pub *Publisher) Close(ctx context.Context) error {
	return writeMessage(ctx, pub.conn, &Message{Type: MessageTypeClose})
}

func (c *Client) Close() error {
	if c.conn == nil {
		return ErrAlreadyClosed
	}
	conn := c.conn
	c.conn = nil
	return conn.CloseNow()
}
