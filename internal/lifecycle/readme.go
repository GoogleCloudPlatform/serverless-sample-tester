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

const (
	// The tag that should appear immediately before code blocks in a README to indicate that the enclosed commands
	// are to be used by this program for building and deploying the sample.
	codeTag = "{sst-run-unix}"

	// A non-quoted backslash in bash at the end of a line indicates a line continuation from the current line to the
	// next line.
	bashLineContChar = '\\'
)

var (
	gcloudCommandRegexp   = regexp.MustCompile(`\bgcloud\b`)
	cloudRunCommandRegexp = regexp.MustCompile(`\brun\b`)

	gcrURLRegexp = regexp.MustCompile(`gcr.io/.+/\S+`)

	mdCodeFenceStartRegexp = regexp.MustCompile("^\\w*`{3,}[^`]*$")

	errNoREADMECommandsFound = fmt.Errorf("[lifecycle.parseREADME]: no commands found")
)

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

		if mdCodeFenceStartRegexp.MatchString(line) {
			if s := scanner.Scan(); !s {
				if err := scanner.Err(); err != nil && !s {
					return nil, fmt.Errorf("[lifecycle.parseREADME] README bufio.Scanner: %w", err)
				}
				return nil, fmt.Errorf("[lifecycle.parseREADME]: unexpected EOF in %s; file ended immediately " +
					"after code tag", filename)
			}

			startCodeBlockLine := scanner.Text()
			m := mdCodeFenceStartRegexp.MatchString(startCodeBlockLine)
			if !m {
				return nil, fmt.Errorf("[lifecycle.parseREADME]: expecting start of code block immediately " +
					"after code tag in %s", filename)
			}

			c := strings.Count(startCodeBlockLine, "`")
			mdCodeFenceEndRegexp := regexp.MustCompile(fmt.Sprintf("^\\w*`{%d,}\\w*$", c))

			var blockClosed bool
			for scanner.Scan() {
				line = scanner.Text()
				if mdCodeFenceEndRegexp.MatchString(line) {
					blockClosed = true
					break
				}

				line = strings.TrimSpace(line)

				// If there is a backslash at the end of the line, this is a multiline command. Keep scanning to get
				// entire command.
				for line[len(line)-1] == bashLineContChar {
					line = line[:len(line)-1]

					if s := scanner.Scan(); !s {
						if err := scanner.Err(); err != nil && !s {
							return nil, fmt.Errorf("[lifecycle.parseREADME] README bufio.Scanner: %w", err)
						}
						return nil, fmt.Errorf("[lifecycle.parseREADME]: unexpected EOF in %s; file ended " +
							"immediately after code tag", filename)
					}

					l := scanner.Text()
					if mdCodeFenceEndRegexp.MatchString(l) {
						return nil, fmt.Errorf("[lifecycle.parseREADME]: unexpected end of code block in %s; " +
							"expecting command line continuation", filename)
					}

					line = line + strings.TrimSpace(l)
				}

				line = os.ExpandEnv(line)
				line = gcrURLRegexp.ReplaceAllString(line, gcrURL)
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
				return nil, fmt.Errorf("[lifecycle.parseREADME]: unexpected EOF in %s; code block not closed",
					filename)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("[lifecycle.parseREADME] README bufio.Scanner: %w", err)
	}

	if len(lifecycle) == 0 {
		return nil, errNoREADMECommandsFound
	}

	return lifecycle, nil
}

// replaceServiceName takes a terminal command string as input and replaces the Cloud Run service name, if any.
// It detects whether the command is a gcloud run command and replaces the last argument that isn't a flag
// with the input service name.
func replaceServiceName(command, serviceName string) string {
	if !(gcloudCommandRegexp.MatchString(command) && cloudRunCommandRegexp.MatchString(command)) {
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
