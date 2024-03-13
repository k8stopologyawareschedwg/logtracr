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
	"testing"
	"time"
)

func TestEmpty(t *testing.T) {
	ft := FakeTime{}
	lc := NewTracrMap(ft.Now)
	sz := lc.Len()
	if sz > 0 {
		t.Errorf("unexpected len > 0: %v", sz)
	}
}

func TestPutGet(t *testing.T) {
	ft := FakeTime{}
	lc := NewTracrMap(ft.Now)
	err := lc.Put("foo", "fizzbuzz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := lc.Get("foo")
	if !ok || got != "fizzbuzz" {
		t.Fatalf("unexpected value: %v (ok=%v)", got, ok)
	}
}

func TestMultiKeyPutGet(t *testing.T) {
	ft := FakeTime{}
	lc := NewTracrMap(ft.Now)
	keys := []string{"foo", "bar", "baz", "buz", "abc"}
	for _, key := range keys {
		err := lc.Put(key, "fizzbuzz")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	for _, key := range keys {
		got, ok := lc.Get(key)
		if !ok || got != "fizzbuzz" {
			t.Fatalf("unexpected value: %v (ok=%v)", got, ok)
		}
	}
	sz := lc.Len()
	if sz != len(keys) {
		t.Errorf("unexpected len: %d expected: %d", sz, len(keys))
	}
}

func TestAppendGet(t *testing.T) {
	ft := FakeTime{}
	lc := NewTracrMap(ft.Now)
	for _, data := range []string{"fizz", "buzz", "fizz", "buzz", "fizz"} {
		err := lc.Put("foo", data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	got, ok := lc.Get("foo")
	if !ok || got != "fizzbuzzfizzbuzzfizz" {
		t.Fatalf("unexpected value: %v (ok=%v)", got, ok)
	}

	sz := lc.Len()
	if sz != 1 {
		t.Errorf("unexpected len: %v", sz)
	}
}

func TestMultiKeyAppendGet(t *testing.T) {
	ft := FakeTime{}
	lc := NewTracrMap(ft.Now)
	keys := []string{"foo", "bar", "baz", "buz", "abc"}
	expected := make(map[string]string)

	for _, data := range []string{"fizz", "buzz", "fizz", "buzz", "fizz"} {
		for _, key := range keys {
			err := lc.Put(key, data+key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expected[key] += data + key
		}
	}
	for _, key := range keys {
		got, ok := lc.Get(key)
		if !ok || got != expected[key] {
			t.Fatalf("unexpected value for key %q: %v (ok=%v)", key, got, ok)
		}
	}

	sz := lc.Len()
	if sz != len(keys) {
		t.Errorf("unexpected len: %d expected: %d", sz, len(keys))
	}
}

type FakeTime struct {
	TS time.Time
}

func (ft FakeTime) Now() time.Time {
	return ft.TS
}
