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

package flusher

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

type BufferAccessor interface {
	GetBuffer(val string) *bytes.Buffer
	PopBuffer(val string) *bytes.Buffer
}

type ErrorPropagation int

const (
	ErrorIgnore = iota
	ErrorPropagate
)

type Config struct {
	BaseDirectory  string
	MaxAge         time.Duration
	ErrPropagation ErrorPropagation
}

type Flusher struct {
	lock sync.Mutex
	cfg  Config
	ba   BufferAccessor
	ages map[string]time.Time
	log  logr.Logger
}

func NewWithLogger(log logr.Logger, cfg Config, ba BufferAccessor) *Flusher {
	return &Flusher{
		cfg:  cfg,
		ba:   ba,
		ages: make(map[string]time.Time),
		log:  log,
	}
}

func (wr *Flusher) MessageDone(val string) {
	wr.lock.Lock()
	defer wr.lock.Unlock()
	wr.ages[val] = time.Now()
}

func (wr *Flusher) FlushAll() {
	now := time.Now()

	wr.lock.Lock()
	defer wr.lock.Unlock()

	flushed := maps.Keys(wr.ages)

	for fl := range flushed {
		buf := wr.ba.GetBuffer(fl)
		if buf == nil {
			if wr.cfg.ErrPropagation == ErrorPropagate {
				wr.log.Info("missing buffer", "ident", fl)
			}
			continue
		}
		err := wr.flushBuffer(fl, buf)
		if err != nil {
			if wr.cfg.ErrPropagation == ErrorPropagate {
				wr.log.Error(err, "cannot store", "ident", fl)
			}
			continue
		}
	}

	for fl := range flushed {
		wr.ages[fl] = now
	}
}

func (wr *Flusher) Flush(now time.Time) []string {
	wr.lock.Lock()
	elapsed := make([]string, 0, len(wr.ages))
	flushed := make([]string, 0, len(wr.ages))

	for val, ts := range wr.ages {
		if now.Sub(ts) < wr.cfg.MaxAge {
			continue
		}
		elapsed = append(elapsed, val)
	}
	wr.lock.Unlock()

	// process elapsed
	for _, el := range elapsed {
		buf := wr.ba.PopBuffer(el)
		if buf == nil {
			if wr.cfg.ErrPropagation == ErrorPropagate {
				wr.log.Info("missing buffer", "ident", el)
			}
			continue
		}
		err := wr.flushBuffer(el, buf)
		if err != nil {
			if wr.cfg.ErrPropagation == ErrorPropagate {
				wr.log.Error(err, "cannot store", "ident", el)
			}
			continue
		}
		flushed = append(flushed, el)
	}

	wr.lock.Lock()
	for _, fl := range flushed {
		// need to check again. What if a late log refreshed the ts?
		ts, ok := wr.ages[fl]
		if ok && now.Sub(ts) < wr.cfg.MaxAge {
			continue
		}
		delete(wr.ages, fl)
	}
	wr.lock.Unlock()

	return flushed
}

func (wr *Flusher) flushBuffer(ident string, buf *bytes.Buffer) (rerr error) {
	ident = strings.ReplaceAll(ident, "/", "_")
	fullPath := filepath.Join(wr.cfg.BaseDirectory, ident)
	var dst *os.File
	dst, rerr = os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if rerr != nil {
		return rerr
	}
	defer func() {
		rerr = dst.Close()
	}()
	_, rerr = dst.WriteString(buf.String())
	if rerr != nil {
		return rerr
	}
	// TODO: flush TS
	return nil
}
