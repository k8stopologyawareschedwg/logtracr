/*
 * Copyright 2024 Red Hat, Inc.
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
	"reflect"
	"testing"
	"time"
)

func TestConfigFromEnv(t *testing.T) {
	type testCase struct {
		name        string
		env         map[string]string
		expectedCfg Config
		expectedErr bool
	}

	testCases := []testCase{
		{
			name:        "empty env",
			expectedErr: true,
		},
		{
			name: "all set",
			env: map[string]string{
				LogTracrDirEnvVar:      "/run/dump",
				LogTracrIntervalEnvVar: "10s",
			},
			expectedCfg: Config{
				DumpInterval:  10 * time.Second,
				DumpDirectory: "/run/dump",
			},
		},
		{
			name: "missing dir",
			env: map[string]string{
				LogTracrIntervalEnvVar: "10s",
			},
			expectedErr: true,
		},
		{
			name: "empty dir",
			env: map[string]string{
				LogTracrDirEnvVar:      "",
				LogTracrIntervalEnvVar: "10s",
			},
			expectedErr: true,
		},
		{
			name: "missing interval",
			env: map[string]string{
				LogTracrDirEnvVar: "/run/dump",
			},
			expectedErr: true,
		},
		{
			name: "malformed interval (1)",
			env: map[string]string{
				LogTracrDirEnvVar:      "/run/dump",
				LogTracrIntervalEnvVar: "10",
			},
			expectedErr: true,
		},
		{
			name: "malformed interval (2)",
			env: map[string]string{
				LogTracrDirEnvVar:      "/run/dump",
				LogTracrIntervalEnvVar: "x",
			},
			expectedErr: true,
		},
		{
			name: "malformed interval (3)",
			env: map[string]string{
				LogTracrDirEnvVar:      "/run/dump",
				LogTracrIntervalEnvVar: "_a",
			},
			expectedErr: true,
		},
		{
			name: "malformed interval (4)",
			env: map[string]string{
				LogTracrDirEnvVar:      "/run/dump",
				LogTracrIntervalEnvVar: "0s",
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for key, val := range tc.env {
				t.Setenv(key, val)
			}
			cfg, err := ConfigFromEnv()
			gotErr := (err != nil)
			if gotErr != tc.expectedErr {
				t.Fatalf("got error %v but expected error %v", gotErr, tc.expectedErr)
			}
			if !tc.expectedErr && !reflect.DeepEqual(cfg, tc.expectedCfg) {
				t.Errorf("got config %+v but expected config %+v", cfg, tc.expectedCfg)
			}
		})
	}
}
