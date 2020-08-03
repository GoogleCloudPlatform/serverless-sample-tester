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
var codeTag = "sst-run-linuxmacos"

// parseREADME parses a README file with the given name. It reads terminal commands surrounded by one of the codeTags
// listed above and loads them into a Lifecycle. In the process, it replaces the Cloud Run service name and Container
// Registry tag with the provided inputs.
func parseREADME(filename, serviceName, gcrURL string) (Lifecycle, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var lifecycle Lifecycle

	inCode := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "```") && inCode {
			inCode = false
		} else if strings.Contains(line, codeTag) {
			scanner.Scan()
			startCodeBlockLine := scanner.Text()

			if !strings.Contains(startCodeBlockLine, "```") {
				return nil, fmt.Errorf("[lifecycle.parseREADME] parsing README: incorrect format")
			}

			inCode = true
		} else if inCode {
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
	}

	if inCode {
		return nil, fmt.Errorf("[lifecycle.parseREADME] parsing README: incorrect format")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lifecycle) == 0 {
		return nil, fmt.Errorf("[lifecycle.parseREADME] parsing README: no commands found")
	}

	return lifecycle, nil
}

// replaceServiceName takes a terminal command string as input and replaces the Cloud Run service name, if any.
// It detects whether the command is a gcloud run command and replaces the last argument that isn't a flag
// with the input service name.
func replaceServiceName(commandStr, serviceName string) string {
	containsGcloud, _ := regexp.MatchString(`\bgcloud\b`, commandStr)
	containsRun, _ := regexp.MatchString(`\brun\b`, commandStr)

	if !(containsGcloud && containsRun) {
		return commandStr
	}

	sp := strings.Split(commandStr, " ")
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
func replaceGCRURL(commandStr string, gcrURL string) string {
	re := regexp.MustCompile(`gcr.io/.+/\S+`)
	return re.ReplaceAllString(commandStr, gcrURL)
}
