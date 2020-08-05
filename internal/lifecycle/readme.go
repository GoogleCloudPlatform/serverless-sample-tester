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

package lifecycle

import (
	"bufio"
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// The tag that should appear immediately before code blocks in a README to indicate that the enclosed commands
// are to be used by this program for building and deploying the sample.
var codeTag = "{sst-run-unix}"

// parseREADME parses a README file with the given name. It reads terminal commands surrounded by one of the codeTags
// listed above and loads them into a Lifecycle. In the process, it replaces the Cloud Run service name and Container
// Registry tag with the provided inputs.
func parseREADME(filename, serviceName, gcrURL string) (Lifecycle, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("[lifecycle.parseREADME] os.Open %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lifecycle Lifecycle
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, codeTag) {
			s := scanner.Scan()
			startCodeBlockLine := scanner.Text()

			c := strings.Contains(startCodeBlockLine, "```")
			if !s || !c {
				if err := scanner.Err(); err != nil && !s {
					return nil, fmt.Errorf("[lifecycle.parseREADME] README bufio.Scanner: %w", err)
				}

				if !c {
					return nil, fmt.Errorf("[lifecycle.parseREADME] parsing README: expecting start of code block immediately after code tag")
				} else { // scanner.Scan falied, but no error found. EOF.
					return nil, fmt.Errorf("[lifecycle.parseREADME] parsing README: unexpected EOF; file ended immediately after code tag")
				}
			}

			var blockClosed bool
			for scanner.Scan() {
				line = scanner.Text()
				if strings.Contains(line, "```") {
					blockClosed = true
					break
				}

				line = strings.TrimSpace(line)

				// If there is a backslash at the end of the line, this is a multiline command. Keep scanning to get
				// entire command.
				for line[len(line)-1] == '\\' {
					line = line[:len(line)-1]

					scanner.Scan()
					line = line + strings.TrimSpace(scanner.Text())
				}

				line = os.ExpandEnv(line)
				line = replaceGCRURL(line, gcrURL)
				line = replaceServiceName(line, serviceName)
				sp := strings.Split(line, " ")

				var cmd *exec.Cmd
				if strings.Contains(line, "gcloud") {
					a := append(util.GcloudCommonFlags, sp[1:]...)
					cmd = exec.Command("gcloud", a...)
				} else {
					cmd = exec.Command(sp[0], sp[1:]...)
				}

				lifecycle = append(lifecycle, cmd)
			}

			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("[lifecycle.parseREADME] README bufio.Scanner: %w", err)
			}

			if !blockClosed {
				return nil, fmt.Errorf("[lifecycle.parseREADME] parsing README: code block not closed before end of file")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("[lifecycle.parseREADME] README bufio.Scanner: %w", err)
	}

	if len(lifecycle) == 0 {
		return nil, fmt.Errorf("[lifecycle.parseREADME] parsing README: no commands found")
	}

	return lifecycle, nil
}

// replaceServiceName takes a terminal command string as input and replaces the Cloud Run service name, if any.
// It detects whether the command is a gcloud run command and replaces the last argument that isn't a flag
// with the input service name.
func replaceServiceName(command, serviceName string) string {
	r1 := regexp.MustCompile(`\bgcloud\b`)
	r2 := regexp.MustCompile(`\brun\b`)

	containsGcloud := r1.MatchString(command)
	containsRun := r2.MatchString(command)

	if !(containsGcloud && containsRun) {
		return command
	}

	sp := strings.Split(command, " ")
	for i := len(sp) - 1; i >= 0; i-- {
		if !strings.Contains(sp[i], "--") {
			sp[i] = serviceName
			break
		}
	}

	return strings.Join(sp, " ")
}

// replaceGCRURL takes a terminal command string as input and replaces the URL of a container image stored in the
// GCP Container Registry with the given URL.
func replaceGCRURL(command string, gcrURL string) string {
	re := regexp.MustCompile(`gcr.io/.+/\S+`)
	return re.ReplaceAllString(command, gcrURL)
}
