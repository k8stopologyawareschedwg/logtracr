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
	"os"
	"strconv"
	"time"
)

const (
	LogTracrDirEnvVar      string = "LOGTRACR_DUMP_DIR"
	LogTracrIntervalEnvVar string = "LOGTRACR_DUMP_INTERVAL"
	LogTracrVerboseEnvVar  string = "LOGTRACR_VERBOSE"
)

type EnvVarMissing struct {
	Name string
}

func (ev *EnvVarMissing) Error() string {
	return fmt.Sprintf("missing env var %q", ev.Name)
}

type UnsupportedSetting struct {
	Name  string
	Value string
}

func (us *UnsupportedSetting) Error() string {
	return fmt.Sprintf("setting %q unsupported value %q", us.Name, us.Value)
}

func ConfigFromEnv() (Config, error) {
	var ok bool
	var err error
	var conf Config

	conf.DumpDirectory, ok = os.LookupEnv(LogTracrDirEnvVar)
	if !ok {
		return conf, &EnvVarMissing{Name: LogTracrDirEnvVar}
	}
	if conf.DumpDirectory == "" {
		return conf, &UnsupportedSetting{Name: LogTracrDirEnvVar, Value: conf.DumpDirectory}
	}

	val, ok := os.LookupEnv(LogTracrIntervalEnvVar)
	if !ok {
		return conf, &EnvVarMissing{Name: LogTracrIntervalEnvVar}
	}
	conf.DumpInterval, err = time.ParseDuration(val)
	if err != nil {
		return conf, fmt.Errorf("cannot parse interval from %v: %w", val, err)
	}
	if conf.DumpInterval == 0 {
		return conf, &UnsupportedSetting{Name: LogTracrIntervalEnvVar, Value: val}
	}

	verb, ok := os.LookupEnv(LogTracrVerboseEnvVar)
	if !ok {
		return conf, &EnvVarMissing{Name: LogTracrVerboseEnvVar}
	}
	conf.Verbose, err = strconv.Atoi(verb)
	if err != nil {
		return conf, &UnsupportedSetting{Name: LogTracrIntervalEnvVar, Value: verb}
	}

	return conf, nil
}
