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
	Verbose       int           `json:"verbose"`
	DumpInterval  time.Duration `json:"dumpInterval"`
	DumpDirectory string        `json:"dumpDirectory"`
}

type Params struct {
	Conf        Config
	Timestamper TimeFunc
}

func SetupWithEnv(ctx context.Context) (logr.Logger, logr.Logger, error) {
	backend := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	conf, err := ConfigFromEnv()
	if err != nil {
		return stdr.New(backend), logr.Discard(), err
	}

	mainSink, flowSink := SetupWithParams(ctx, backend, Params{
		Conf:        conf,
		Timestamper: time.Now,
	})
	return mainSink, flowSink, nil
}

func SetupWithParams(ctx context.Context, backend *log.Logger, params Params) (logr.Logger, logr.Logger) {
	mainSink := stdr.New(backend)
	mainSink.Info("starting", "configuration", toJSON(params.Conf))

	traces := NewTracrMap(params.Timestamper)
	flowSink := New(backend, traces, params.Conf.Verbose, stdr.Options{}) // promote to final sink
	go RunForever(ctx, flowSink, params.Conf.DumpInterval, params.Conf.DumpDirectory, traces)

	return mainSink, flowSink
}

func toJSON(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return "<ERROR>"
	}
	return string(data)
}
