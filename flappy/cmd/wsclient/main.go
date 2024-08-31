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
	"fmt"
	"log"
	"os"

	"github.com/eiannone/keyboard"
	"github.com/syumai/ebiten-playground/flappy/remote"
)

func realMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := remote.NewClient(ctx)
	if err != nil {
		return err
	}
	pub, err := client.NewPublisher(ctx)
	if err != nil {
		return err
	}

	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()

	fmt.Println("Press Space to publish message")
	for {
		select {
		case <-ctx.Done():
			break
		default:
		}

		_, key, err := keyboard.GetKey()
		if err != nil {
			return err
		}
		if key == keyboard.KeySpace {
			log.Println("published message")
			if err := pub.Publish(ctx, &remote.Message{Type: remote.MessageTypePublish}); err != nil {
				return err
			}
		}
		if key == keyboard.KeyCtrlC {
			break
		}
	}
	if err := pub.Publish(ctx, &remote.Message{Type: remote.MessageTypeClose}); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
