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

package fanout

import (
	"sync"

	"github.com/go-logr/logr"
)

type Fanout struct {
	// protects the `leaves` slice, not the logger instances
	rwlock sync.RWMutex
	leaves []logr.Logger
}

func (fo *Fanout) GetUnderlying() []logr.Logger {
	return fo.leaves
}

func NewWithLeaves(leaves ...logr.Logger) logr.LogSink {
	return &Fanout{
		leaves: leaves,
	}
}

// Init is not implemented and does not use any runtime info.
func (fo *Fanout) Init(info logr.RuntimeInfo) {
	// not implemented
}

// Enabled delegate this check to leaves later on.
// There's no easy way to query loggers (vs their logsinks, but
// we don't have access to them), so this is the best we can do
func (fo *Fanout) Enabled(level int) bool {
	return true
}

// Info dispatches the call to all the leaves
func (fo *Fanout) Info(level int, msg string, kv ...any) {
	fo.rwlock.RLock() // we don't mutate the leaves slice
	defer fo.rwlock.RUnlock()
	for _, leaf := range fo.leaves {
		leaf.Info(msg, kv...)
	}
}

// Error dispatches the call to all the leaves
func (fo *Fanout) Error(err error, msg string, kv ...any) {
	fo.rwlock.RLock() // we don't mutate the leaves slice
	defer fo.rwlock.Unlock()
	for _, leaf := range fo.leaves {
		leaf.Error(err, msg, kv...)
	}
}

// WithValues dispatches the call to all the leaves, mutating them
func (fo *Fanout) WithValues(kv ...any) logr.LogSink {
	fo.rwlock.Lock()
	defer fo.rwlock.Unlock()
	leaves := make([]logr.Logger, 0, len(fo.leaves))
	for _, leaf := range fo.leaves {
		leaves = append(leaves, leaf.WithValues(kv...))
	}
	return &Fanout{
		leaves: leaves,
	}
}

// WithName dispatches the call to all the leaves, mutating them
func (fo *Fanout) WithName(name string) logr.LogSink {
	fo.rwlock.Lock()
	defer fo.rwlock.Unlock()
	leaves := make([]logr.Logger, 0, len(fo.leaves))
	for _, leaf := range fo.leaves {
		leaves = append(leaves, leaf.WithName(name))
	}
	return &Fanout{
		leaves: leaves,
	}
}

// Underlier exposes access to the underlying logging function. Since
// callers only have a logr.Logger, they have to know which
// implementation is in use, so this interface is less of an
// abstraction and more of a way to test type conversion.
type Underlier interface {
	GetUnderlying() []logr.Logger
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &Fanout{}
var _ Underlier = &Fanout{}

// TODO: implement logr.CallDepthLogSink
