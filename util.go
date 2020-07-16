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

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const cloudRunServiceNameRandSuffixLen = 10

// gcloudCommandBuildcreates an exec.Cmd that calls the external gcloud executable. The exec.Cmd's name is set to
// "gcloud" and the args are the ones provided in addition to a project flag and quiet flag.
func gcloudCommandBuild(arg []string) *exec.Cmd {
	var gcpProject string
	if s == nil {
		gcpProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	} else {
		gcpProject = s.projectID
	}

	arg = append(arg, fmt.Sprintf("--project=%s", gcpProject), "--quiet")
	cmd := exec.Command("gcloud", arg...)

	return cmd
}

// execCommand executes an exec.Cmd. It redirects the commands stderr to this program's stderr and returns the output
// in the form of a string.
func execCommand(cmd *exec.Cmd) string {
	if s == nil {
		cmd.Dir = sampleDir
	} else {
		cmd.Dir = s.dir
	}

	cmd.Stderr = os.Stderr

	log.Println("Executing ", cmd.String())

	b, err := cmd.Output()
	if err != nil {
		fmt.Println(string(b))
		log.Panicf("Error with exec cmd: %v\n", err)
	}

	result := string(b)
	return strings.TrimSpace(result)
}
