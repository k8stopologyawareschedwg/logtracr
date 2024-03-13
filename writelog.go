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
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

func RunForever(ctx context.Context, logger logr.Logger, interval time.Duration, baseDirectory string, lc *TracrMap) {
	// let's try to keep the amount of code we do in init() at minimum.
	// This may happen if the container didn't have the directory mounted
	discard := !existsBaseDirectory(baseDirectory)
	if discard {
		logger.Info("base directory not found, will discard everything", "baseDirectory", baseDirectory)
	}

	delta := interval - 10*time.Millisecond // TODO
	logger.Info("dump loop info", "interval", interval, "delta", delta)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			// keep the size at bay by popping old data even if we just discard it
			expireds := lc.PopExpired(now, delta)
			for _, expired := range expireds {
				if discard {
					continue
				}
				// intentionally swallow error.
				// - if we hit issues and we do log (V=2, like), we will clog the regular log
				// - if we hit issues and we do NOT log (v=5, like) we will not see it anyway
				writeLogTrace(baseDirectory, expired.logID, expired.data)
			}
			if len(expireds) > 0 {
				logger.V(4).Info("processed logs", "entries", len(expireds), "stored", !discard)
			}
		}
	}
}

func writeLogTrace(statusDir string, logName string, data bytes.Buffer) error {
	logName = fixLogName(logName)
	dst, err := os.OpenFile(filepath.Join(statusDir, logName+".log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	if _, err := dst.Write(data.Bytes()); err != nil {
		dst.Close() // swallow error because we want to bubble up the write error
		return err
	}
	return dst.Close()
}

func existsBaseDirectory(baseDir string) bool {
	info, err := os.Stat(baseDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fixLogName(name string) string {
	return strings.ReplaceAll(name, "/", "__")
}
