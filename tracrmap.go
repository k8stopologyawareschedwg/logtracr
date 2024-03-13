/*
 * Copyright 2023 Red Hat, Inc.
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
	"bytes"
	"sync"
	"time"
)

type SpanBlob struct {
	logID      string
	data       bytes.Buffer
	lastUpdate time.Time
}

func (ln *SpanBlob) IsExpired(now time.Time, delta time.Duration) bool {
	return now.Sub(ln.lastUpdate) >= delta
}

type TimeFunc func() time.Time

type TracrMap struct {
	mutex sync.Mutex
	// map logID -> data
	spanBlobs map[string]*SpanBlob
	timeFunc  TimeFunc
}

func NewTracrMap(timeFunc TimeFunc) *TracrMap {
	return &TracrMap{
		spanBlobs: make(map[string]*SpanBlob),
		timeFunc:  timeFunc,
	}
}

func (lc *TracrMap) Put(logID, data string) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	buf := lc.bufferFor(logID)
	_, err := buf.WriteString(data)
	return err
}

func (lc *TracrMap) PopExpired(now time.Time, delta time.Duration) []*SpanBlob {
	ret := []*SpanBlob{}
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	for logID, SpanBlob := range lc.spanBlobs {
		if !SpanBlob.IsExpired(now, delta) {
			continue
		}
		ret = append(ret, SpanBlob)
		delete(lc.spanBlobs, logID)
	}
	return ret
}

// Get is (mostly) meant for testing purposes
func (lc *TracrMap) Get(logID string) (string, bool) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	node, ok := lc.spanBlobs[logID]
	if !ok {
		return "", false
	}
	return node.data.String(), true
}

// Len is (mostly) meant for testing purposes
func (lc *TracrMap) Len() int {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	return len(lc.spanBlobs)
}

func (lc *TracrMap) bufferFor(logID string) *bytes.Buffer {
	node, ok := lc.spanBlobs[logID]
	if !ok {
		node = &SpanBlob{
			logID: logID,
		}
		lc.spanBlobs[logID] = node
	}
	node.lastUpdate = lc.timeFunc()
	return &node.data
}
