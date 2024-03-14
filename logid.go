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

var (
	logIDKey = "logID"
)

func SetLogIDKey(val string) {
	logIDKey = val
}

func GetLogIDKey() string {
	return logIDKey
}

func findLogID(values []any) (string, bool) {
	if len(values) < 2 || len(values)%2 != 0 {
		return "", false // should never happen
	}
	for i := 0; i < len(values); i += 2 {
		kv, ok := values[i].(string)
		if !ok || kv != logIDKey {
			continue
		}
		vv, ok := values[i+1].(string)
		return vv, ok
	}
	return "", false
}
