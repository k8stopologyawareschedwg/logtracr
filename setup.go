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
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

type Config struct {
	DumpInterval  time.Duration `json:"dumpInterval"`
	DumpDirectory string        `json:"dumpDirectory"`
}

type Params struct {
	Conf        Config
	Timestamper TimeFunc
}

func SetupWithEnv(ctx context.Context) (logr.Logger, error) {
	conf, err := ConfigFromEnv()
	if err != nil {
		return logr.Discard(), err
	}
	return SetupWithParams(ctx, Params{
		Conf:        conf,
		Timestamper: time.Now,
	}), nil
}

func SetupWithParams(ctx context.Context, params Params) logr.Logger {
	backend := stdr.New(log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile))
	backend.Info("starting", "configuration", toJSON(params.Conf))
	traces := NewTracrMap(params.Timestamper)
	lh := New(traces, stdr.Options{}) // promote to final lh
	go RunForever(ctx, backend, params.Conf.DumpInterval, params.Conf.DumpDirectory, traces)
	return lh
}

func toJSON(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return "<ERROR>"
	}
	return string(data)
}
