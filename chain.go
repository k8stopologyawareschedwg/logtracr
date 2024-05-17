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
	"github.com/go-logr/logr"
	"github.com/wojas/genericr"
)

var (
	backend    logr.Logger
	registered bool
)

func RegisterBackend(lh logr.Logger) {
	backend = lh
	registered = true
}

func Tee(lh logr.Logger) logr.Logger {
	if !registered {
		return lh
	}
	return Chain(backend, lh)
}

func Chain(lhs ...logr.Logger) logr.Logger {
	return logr.New(genericr.New(func(e genericr.Entry) {
		for _, lh := range lhs {
			if e.Error != nil {
				lh.Error(e.Error, e.Message, e.Fields...)
			} else {
				lh.Info(e.Message, e.Fields...)
			}
		}
	}))
}
