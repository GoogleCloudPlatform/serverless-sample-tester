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
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// The directory in which all of the exec.Cmds will be run.
var commandsDir string

// SetCommandsDir is the setter method for commandsDir.
func SetCommandsDir(d string) {
	commandsDir = d
}

// GcloudCommandBuild creates an exec.Cmd that calls the external gcloud executable. The exec.Cmd's name is set to
// "gcloud" and the args are the ones provided in addition to a project flag and quiet flag.
func GcloudCommandBuild(arg ...string) *exec.Cmd {
	arg = append(arg, "--quiet")
	cmd := exec.Command("gcloud", arg...)

	return cmd
}

// ExecCommand executes an exec.Cmd. It redirects the command's stderr to this program's stderr and returns the output
// in the form of a string. The command will be run in commandsDir.
func ExecCommand(cmd *exec.Cmd) (string, error) {
	cmd.Dir = commandsDir
	cmd.Stderr = os.Stderr

	log.Printf("Executing %v\n", cmd)

	b, err := cmd.Output()
	output := strings.TrimSpace(string(b))
	if err != nil {
		fmt.Println(output)
		return output, fmt.Errorf("[util.ExecCommand] error executing external command %v: %w", cmd, err)
	}

	return output, nil
}
