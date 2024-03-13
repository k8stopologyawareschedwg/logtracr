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
	"os"
	"path/filepath"
	"testing"
)

func TestWriteLogTraceAppendsData(t *testing.T) {
	dir, err := os.MkdirTemp("", "logtracr-dump-data")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	buf1 := bytes.NewBufferString("fizzbuzz\n")
	err = writeLogTrace(dir, "foo-test", *buf1)
	if err != nil {
		t.Fatal(err)
	}

	buf2 := bytes.NewBufferString("foobar\n")
	err = writeLogTrace(dir, "foo-test", *buf2)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "foo-test.log"))
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	expected := "fizzbuzz\nfoobar\n"
	if got != expected {
		t.Errorf("read error\ngot=[%s]\nexp=[%s]", got, expected)
	}
}
