/*
 * Copyright 2025 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logtracr

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	"github.com/k8stopologyawareschedwg/logtracr/demuxer"
	"github.com/k8stopologyawareschedwg/logtracr/fanout"
	"github.com/k8stopologyawareschedwg/logtracr/flusher"
)

type Config struct {
	LogKey      string
	FlushPeriod time.Duration
	Flusher     flusher.Config
}

type CancelFunc func()
type AsyncFlushFunc func()

type Control struct {
	Cancel     CancelFunc
	AsyncFlush AsyncFlushFunc
}

func NewWithConfig(lh logr.Logger, cfg Config) (logr.Logger, Control) {
	dmx := demuxer.NewWithOptions(&demuxer.Options{
		KeyFinder: demuxer.GenericKeyFinder(cfg.LogKey),
	})

	fl := flusher.NewWithLogger(lh, cfg.Flusher, dmx)

	dmx.Register(fl.MessageDone)

	ctx, cancel := context.WithCancel(context.Background())

	// TODO close on cancel?
	flushReqCh := make(chan struct{})

	go flushLoop(ctx, flushReqCh, cfg.FlushPeriod, fl)

	return logr.New(fanout.NewWithLeaves(lh, logr.New(dmx))), Control{
		Cancel: CancelFunc(cancel),
		AsyncFlush: func() {
			flushReqCh <- struct{}{}
		},
	}
}

func flushLoop(ctx context.Context, flushReqCh chan struct{}, flushPeriod time.Duration, fl *flusher.Flusher) {
	ticker := time.NewTicker(flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fl.FlushAll()
			return
		case <-flushReqCh:
			fl.FlushAll()
		case ts := <-ticker.C:
			fl.Flush(ts)
		}
	}
}
