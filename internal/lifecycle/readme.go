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
	mdCodeFenceStartRegexp = regexp.MustCompile("^\\w*`{3,}[^`]*$")

	errNoReadmeCodeBlocksFound   = fmt.Errorf("lifecycle.extractCodeBlocks: no code blocks immediately preceded by %s found", codeTag)
	errCodeBlockNotClosed        = fmt.Errorf("unexpected EOF: code block not closed")
	errCodeBlockStartNotFound    = fmt.Errorf("expecting start of code block immediately after code tag")
	errEOFAfterCodeTag           = fmt.Errorf("unexpected EOF: file ended immediately after code tag")
	errCodeBlockEndAfterLineCont = "end of code block: expecting command line continuation"
)

// codeBlock is a slice of strings containing terminal commands. codeBlocks, for example, could be used to hold the
// terminal commands inside of a Markdown code block.
type codeBlock []string

// toCommands extracts the terminal commands contained within the current codeBlock. It handles the expansion of
// environment variables and line continuations. It also detects Cloud Run service names Google Container Registry
// container image URLs and replaces them with the ones provided.
func (cb codeBlock) toCommands(serviceName, gcrURL string) ([]*exec.Cmd, error) {
	var cmds []*exec.Cmd

	for i := 0; i < len(cb); i++ {
		line := cb[i]
		if line == "" {
			continue
		}

		// If there is a backslash at the end of the line, this is a multiline command. Keep scanning to get entire
		// command.
		for line[len(line)-1] == bashLineContChar {
			line = line[:len(line)-1]

			i++
			if i >= len(cb) {
				return nil, fmt.Errorf("%s; code block dump:\n%s", errCodeBlockEndAfterLineCont, strings.Join(cb, "\n"))
			}

			l := cb[i]
			if l == "" {
				break
			}

			line = line + l
		}

		line = os.ExpandEnv(line)
		line = gcrURLRegexp.ReplaceAllString(line, gcrURL)

		sp := strings.Split(line, " ")

		err := replaceServiceName(sp[0], sp[1:], serviceName)
		if err != nil {
			return nil, fmt.Errorf("lifecycle.replaceServiceName: %s: %w", line, err)
		}

		var cmd *exec.Cmd
		if sp[0] == "gcloud" {
			a := append(util.GcloudCommonFlags, sp[1:]...)
			cmd = exec.Command("gcloud", a...)
		} else {
			cmd = exec.Command(sp[0], sp[1:]...)
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// parseREADME parses a README file with the given name. It parses terminal commands in code blocks annotated by the
// codeTag and loads them into a Lifecycle. In the process, it replaces the Cloud Run service name and Container
// Registry tag with the provided inputs. It also expands environment variables and supports bash-style line
// continuations.
func parseREADME(filename, serviceName, gcrURL string) (Lifecycle, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	return extractLifecycle(scanner, serviceName, gcrURL)
}

// extractLifecycle is a helper function for parseREADME. It takes a scanner that reads from a Markdown file and parses
// terminal commands in code blocks annotated by the codeTag and loads them into a Lifecycle. In the process, it
// replaces the Cloud Run service name and Container Registry tag with the provided inputs. It also expands environment
// variables and supports bash-style line continuations.
func extractLifecycle(scanner *bufio.Scanner, serviceName, gcrURL string) (Lifecycle, error) {
	codeBlocks, err := extractCodeBlocks(scanner)
	if err != nil {
		return nil, fmt.Errorf("lifecycle.extractCodeBlocks: %w", err)
	}

	if len(codeBlocks) == 0 {
		return nil, errNoReadmeCodeBlocksFound
	}

	var l Lifecycle
	for _, b := range codeBlocks {
		cmds, err := b.toCommands(serviceName, gcrURL)
		if err != nil {
			return l, fmt.Errorf("codeBlock.toCommands: %w", err)
		}

		l = append(l, cmds...)
	}

	return l, nil
}

// codeBlocks extracts code blocks out of a bufio.Scanner that's reading from a Markdown file immediately prefaced with
// a line containing codeTag. It returns an 2d slice of code blocks, each containing an array of lines contained within
// that code block.
func extractCodeBlocks(scanner *bufio.Scanner) ([]codeBlock, error) {
	var blocks []codeBlock

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.Contains(line, codeTag) {
			if s := scanner.Scan(); !s {
				if err := scanner.Err(); err != nil {
					return nil, fmt.Errorf("line %d: bufio.Scanner.Scan: %w", lineNum, err)
				}
				return nil, errEOFAfterCodeTag
			}
			lineNum++

			startCodeBlockLine := scanner.Text()
			m := mdCodeFenceStartRegexp.MatchString(startCodeBlockLine)
			if !m {
				return nil, fmt.Errorf("line %d: %w", lineNum, errCodeBlockStartNotFound)
			}

			c := strings.Count(startCodeBlockLine, "`")
			mdCodeFenceEndRegexp := regexp.MustCompile(fmt.Sprintf("^\\w*`{%d,}\\w*$", c))

			var block codeBlock
			var blockClosed bool
			for scanner.Scan() {
				lineNum++
				line = strings.TrimSpace(scanner.Text())
				if mdCodeFenceEndRegexp.MatchString(line) {
					blockClosed = true
					break
				}

				block = append(block, line)
			}

			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("line %d: bufio.Scanner.Scan: %w", lineNum, err)
			}

			if !blockClosed {
				return nil, errCodeBlockNotClosed
			}

			blocks = append(blocks, block)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("line %d: bufio.Scanner.Scan: %w", lineNum, err)
	}

	return blocks, nil
}
