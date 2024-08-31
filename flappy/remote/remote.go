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
	"github.com/coder/websocket"
)

const WSServerPort = "7382"

var ErrAlreadyClosed = errors.New("already closed")

type Remote struct {
	conn *websocket.Conn
}

func NewRemote(ctx context.Context) (*Remote, error) {
	c, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://localhost:%s", WSServerPort), nil)
	if err != nil {
		return nil, err
	}
	return &Remote{conn: c}, nil
}

func (r *Remote) Subscribe() error {

}

func (r *Remote) Close() error {
	if r.conn == nil {
		return ErrAlreadyClosed
	}
	c := r.conn
	r.conn = nil
	return c.CloseNow()
}
