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
	"fmt"
	"testing"
)

func TestFindLogID(t *testing.T) {
	type testCase struct {
		name        string
		values      []any
		expectedVal string
		expectedOK  bool
	}

	testCases := []testCase{
		{
			name: "empty",
		},
		{
			name:   "odd values",
			values: []any{"foo"},
		},
		{
			name:   "long list, non matching",
			values: makeArgs(1024),
		},
		{
			name:   "odd values, matching",
			values: []any{GetLogIDKey()},
		},
		{
			name:   "minimal match, wrong type",
			values: []any{GetLogIDKey(), 123},
		},
		{
			name:        "minimal match",
			values:      []any{GetLogIDKey(), "test123"},
			expectedVal: "test123",
			expectedOK:  true,
		},
		{
			name:        "match not in first position",
			values:      []any{"foo", "bar", "baz", 2.2, GetLogIDKey(), "testABC", "xyz", 123},
			expectedVal: "testABC",
			expectedOK:  true,
		},
		{
			name:   "match not in first position, misaligned",
			values: []any{"foo", "bar", "baz", GetLogIDKey(), "testABC", "xyz", 123},
		},
		{
			name:        "long list, first position",
			values:      append([]any{GetLogIDKey(), "testLL01"}, makeArgs(250)...),
			expectedVal: "testLL01",
			expectedOK:  true,
		},
		{
			name:        "long list, last position",
			values:      append(makeArgs(250), GetLogIDKey(), "testLLZZ"),
			expectedVal: "testLLZZ",
			expectedOK:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := findLogID(tc.values)
			if ok != tc.expectedOK {
				t.Fatalf("expected ok=%v got ok=%v", tc.expectedOK, got)
			}
			if got != tc.expectedVal {
				t.Errorf("expected logID value %q got %q", tc.expectedVal, got)
			}
		})
	}
}

func makeArgs(count int) []any {
	ret := []any{}
	for i := 0; i < count; i++ {
		ret = append(ret, fmt.Sprintf("bogus%03d", i))
	}
	return ret
}
