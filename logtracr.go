/*
 * Copyright 2019 The logr Authors.
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
 *
 * Derived from https://github.com/go-logr/stdr/blob/v1.2.2/stdr.go
 */

package logtracr

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/go-logr/stdr"
)

type logtracr struct {
	funcr.Formatter
	std      stdr.StdLogger
	traces   *TracrMap
	logID    string
	hasLogID bool
	verbose  int
}

func New(std stdr.StdLogger, lc *TracrMap, verbose int, opts stdr.Options) logr.Logger {
	sl := &logtracr{
		Formatter: funcr.NewFormatter(funcr.Options{
			LogCaller: funcr.MessageClass(opts.LogCaller),
		}),
		std:     std,
		traces:  lc,
		verbose: verbose,
	}

	// For skipping our own logger.Info/Error.
	sl.Formatter.AddCallDepth(1 + opts.Depth)

	return logr.New(sl)
}

func (l logtracr) Enabled(level int) bool {
	return l.verbose >= level || l.logID != ""
}

func (l logtracr) Info(level int, msg string, kvList ...interface{}) {
	prefix, args := l.FormatInfo(level, msg, kvList)
	if prefix != "" {
		args = prefix + ": " + args
	}
	l.store(args, kvList)
	// we can end up here because either we have enough verbosiness
	// OR because stored logID. So we must redo this check.
	if l.verbose < level {
		return
	}
	_ = l.std.Output(l.Formatter.GetDepth()+1, args)
}

func (l logtracr) Error(err error, msg string, kvList ...interface{}) {
	prefix, args := l.FormatError(err, msg, kvList)
	if prefix != "" {
		args = prefix + ": " + args
	}
	l.store(args, kvList)
	_ = l.std.Output(l.Formatter.GetDepth()+1, args)
}

func (l logtracr) WithName(name string) logr.LogSink {
	l.Formatter.AddName(name)
	return &l
}

func (l logtracr) WithValues(kvList ...interface{}) logr.LogSink {
	l.logID, l.hasLogID = findLogID(kvList)
	l.Formatter.AddValues(kvList)
	return &l
}

func (l logtracr) WithCallDepth(depth int) logr.LogSink {
	l.Formatter.AddCallDepth(depth)
	return &l
}

func (l logtracr) store(args string, kvList ...interface{}) {
	if l.traces == nil {
		return
	}
	logID, ok := l.logID, l.hasLogID
	if !ok {
		logID, ok = startsWithLogID(kvList...)
	}
	if !ok {
		return
	}
	l.traces.Put(logID, args) // ignore error intentionally
}

var _ logr.LogSink = &logtracr{}
var _ logr.CallDepthLogSink = &logtracr{}
