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
	"os"
	"os/signal"
	"syscall"

	"github.com/syumai/ebiten-playground/flappy/remote"
	"golang.org/x/sync/errgroup"
)

func main() {
	s := newServer(remote.WSServerPort)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(s.start)
	eg.Go(func() error {
		<-egCtx.Done()
		return s.stop()
	})

	if err := eg.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
