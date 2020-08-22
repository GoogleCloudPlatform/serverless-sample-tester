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
	"errors"
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var gcrURLRegexp = regexp.MustCompile(`gcr.io/.+/\S+`)

// Lifecycle is a list of ordered exec.Cmd that should be run to execute a certain process.
type Lifecycle []*exec.Cmd

// Execute executes the commands of a lifecycle in the provided directory.
func (l Lifecycle) Execute(commandsDir string) error {
	for _, c := range l {
		if c == nil {
			continue
		}

		_, err := util.ExecCommand(c, commandsDir)
		if err != nil {
			return fmt.Errorf("executing Lifecycle command: %w", err)
		}
	}

	return nil
}

// NewLifecycle tries to parse the different options provided for build and deploy command configuration. If none of
// those options are set up, it falls back to reasonable defaults based on whether the sample is java-based
// (has a pom.xml) that doesn't have a Dockerfile or isn't. Also returns a function that cleans up any created local
// resources (e.g. temp files) created while making creating this Lifecycle. This function should be called after this
// Lifecycle is done executing.
func NewLifecycle(sampleDir, serviceName, gcrURL, runRegion string, cloudBuildConfSubs map[string]string) (Lifecycle, func(), error) {
	// First try Cloud Build Config file
	cloudBuildConfigPath := fmt.Sprintf("%s/cloudbuild.yaml", sampleDir)

	if _, err := os.Stat(cloudBuildConfigPath); err == nil {
		lifecycle, cleanup, err := getCloudBuildConfigLifecycle(cloudBuildConfigPath, serviceName, gcrURL, runRegion, cloudBuildConfSubs)
		if err == nil {
			log.Println("Using cloud build config file")
			return lifecycle, cleanup, nil
		}

		return nil, nil, fmt.Errorf("lifecycle.getCloudBuildConfigLifecycle: %s: %w\n", cloudBuildConfigPath, err)
	}

	// Then try README parsing
	var readmePath string
	// Searching for config file
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Config file found, using specified location for README")
		readmePath, _ = filepath.Abs(filepath.Join(sampleDir, viper.GetString("readme")))
	} else {
		log.Println("No config file found, using root directory for README location")
		readmePath = filepath.Join(sampleDir, "README.md")
	}

	if _, err := os.Stat(readmePath); err == nil {
		lifecycle, err := parseREADME(readmePath, serviceName, gcrURL)
		// Show README location
		log.Println("README.md location: " + readmePath)
		if err == nil {
			log.Println("Using build and deploy commands found in README.md")
			return lifecycle, nil, nil
		}

		if !errors.Is(err, errNoReadmeCodeBlocksFound) {
			return nil, nil, fmt.Errorf("lifecycle.parseREADME: %s: %w", readmePath, err)
		}

		log.Printf("No code blocks immediately preceded by %s found in README.md\n", codeTag)
	} else {
		log.Println("No README.md found")
	}

	// Finally fall back to reasonable defaults
	pomPath := filepath.Join(sampleDir, "pom.xml")
	dockerfilePath := filepath.Join(sampleDir, "Dockerfile")

	_, err := os.Stat(pomPath)
	pomE := err == nil

	_, err = os.Stat(dockerfilePath)
	dockerfileE := err == nil

	if pomE && !dockerfileE {
		log.Println("Using default build and deploy commands for java samples without a Dockerfile")
		return buildDefaultJavaLifecycle(serviceName, gcrURL), nil, nil
	}

	log.Println("Using default build and deploy commands for non-java samples or java samples with a Dockerfile")
	return buildDefaultLifecycle(serviceName, gcrURL), nil, nil
}

// buildDefaultLifecycle builds a build and deploy command lifecycle with reasonable defaults for a non-Java
// project. It uses `gcloud builds submit` for building the samples container image and submitting it to the container
// and `gcloud run deploy` for deploying it to Cloud Run.
func buildDefaultLifecycle(serviceName, gcrURL string) Lifecycle {
	a0 := append(util.GcloudCommonFlags, "builds", "submit", fmt.Sprintf("--tag=%s", gcrURL))
	a1 := append(util.GcloudCommonFlags, "run", "deploy", serviceName, fmt.Sprintf("--image=%s", gcrURL),
		"--platform=managed")

	return Lifecycle{
		exec.Command("gcloud", a0...),
		exec.Command("gcloud", a1...),
	}
}

// buildDefaultJavaLifecycle builds a build and deploy command lifecycle with reasonable defaults for Java
// samples. It uses `com.google.cloud.tools:jib-maven-plugin:2.0.0:build` for building the samples container image and
// submitting it to the container and `gcloud run deploy` for deploying it to Cloud Run.
func buildDefaultJavaLifecycle(serviceName, gcrURL string) Lifecycle {
	l := buildDefaultLifecycle(serviceName, gcrURL)

	l[0] = exec.Command("mvn",
		"compile",
		"com.google.cloud.tools:jib-maven-plugin:2.0.0:build",
		fmt.Sprintf("-Dimage=%s", gcrURL),
	)

	return l
}

// replaceServiceName takes a terminal command string as input and replaces the Cloud Run service name, if any.
// If the user specified the service name in $CLOUD_RUN_SERVICE_NAME, it replaces that. Otherwise, as a failsafe,
// it detects whether the command is a gcloud run command and replaces the last argument that isn't a flag
// with the input service name.
func replaceServiceName(name string, args []string, serviceName string) error {
	if !strings.Contains(name, "gcloud") {
		return nil
	}

	var runCmd bool

	// Detects if the user specified the Cloud Run service name in an environment variable
	for i := 0; i < len(args); i++ {
		if args[i] == os.ExpandEnv("$CLOUD_RUN_SERVICE_NAME") {
			args[i] = serviceName
			return nil
		}

		if args[i] == "run" {
			runCmd = true
			break
		}
	}

	if !runCmd {
		return nil
	}

	// Searches for specific gcloud keywords and takes service name from them
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "deploy" || args[i] == "update" {
			args[i+1] = serviceName
			return nil
		}
	}

	// Provides a failsafe if neither of the above options work
	for i := len(args) - 1; i >= 0; i-- {
		if !strings.Contains(args[i], "--") {
			args[i] = serviceName
			break
		}
	}
	return nil
}
