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
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"log"
	"os"
	"os/exec"
)

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
			return fmt.Errorf("[Lifecycle.Execute] executing lifecycle command: %w", err)
		}
	}

	return nil
}


// NewLifecycle tries to parse the different options provided for build and deploy command configuration. If none of
// those options are set up, it falls back to reasonable defaults based on whether the sample is java-based
// (has a pom.xml) that doesn't have a Dockerfile or isn't.
func NewLifecycle(sampleDir, serviceName, gcrURL string) Lifecycle {
	readmePath := fmt.Sprintf("%s/README.md", sampleDir)

	_, err := os.Stat(readmePath)
	readmeExists := err == nil

	if readmeExists {
		lifecycle, err := parseREADME(readmePath, serviceName, gcrURL)
		if err == nil {
			log.Println("Using build and deploy commands found in README")
			return lifecycle
		}
		fmt.Printf("Failed parsing README: %v\n", err)
	}

	pomPath := fmt.Sprintf("%s/pom.xml", sampleDir)
	dockerfilePath := fmt.Sprintf("%s/Dockerfile", sampleDir)

	_, err = os.Stat(pomPath)
	pomExists := err == nil

	_, err = os.Stat(dockerfilePath)
	dockerfileExists := err == nil

	if pomExists && !dockerfileExists {
		log.Println("Using default build and deploy commands for java samples without a Dockerfile")
		return buildDefaultJavaLifecycle(serviceName, gcrURL)
	}

	log.Println("Using default build and deploy commands for non-java samples or java samples with a Dockerfile")
	return buildDefaultLifecycle(serviceName, gcrURL)
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
