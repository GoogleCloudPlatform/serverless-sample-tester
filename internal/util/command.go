// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

// GcloudCommonFlags is a slice of common flags that should be added as arguments to all executions of the external
// gcloud command.
var GcloudCommonFlags = []string{
	"--quiet",
}

// ExecCommand executes an exec.Cmd. If the command exits successfully, its stdout will be returned. If there's an
// error, the command's combined stdout and stderr will be returned in an error. The command will be run in the provided
// directory.
func ExecCommand(cmd *exec.Cmd, dir string) (string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	var stdcombined bytes.Buffer

	cmd.Dir = dir

	cmd.Stdout = io.MultiWriter(&stdout, &stdcombined)
	cmd.Stderr = io.MultiWriter(&stderr, &stdcombined)

	log.Printf("Executing %v\n", cmd)

	err := cmd.Run()
	if err != nil {
		out := strings.TrimSpace(string(stdcombined.Bytes()))
		return "", fmt.Errorf("[util.ExecCommand] error executing external command %v:\n%s\n%w", cmd, out, err)
	}

	out := strings.TrimSpace(string(stdout.Bytes()))
	return out, nil
}
