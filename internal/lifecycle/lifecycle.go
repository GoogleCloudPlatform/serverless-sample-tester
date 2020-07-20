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
	"os"
	"os/exec"
)

// Lifecycle is a list of ordered exec.Cmd that should be run to execute a certain process.
type Lifecycle []*exec.Cmd

// Execute executes the commands of a lifecycle.
func (l Lifecycle) Execute() error {
	for _, c := range l {
		_, err := util.ExecCommand(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetLifecycle returns a lifecycle built with reasonable defaults based on whether the sample is java-based
// (has a pom.xml) that doesn't have a Dockerfile or isn't.
func GetLifecycle(sampleDir, serviceName, gcrURL string) Lifecycle {
	pomPath := fmt.Sprintf("%s/pom.xml", sampleDir)
	dockerfilePath := fmt.Sprintf("%s/Dockerfile", sampleDir)

	_, err := os.Stat(pomPath)
	pomExists := err == nil

	_, err = os.Stat(dockerfilePath)
	dockerfileExists := err == nil

	if pomExists && !dockerfileExists {
		return buildDefaultJavaLifecycle(serviceName, gcrURL)
	}

	return buildDefaultLifecycle(serviceName, gcrURL)
}

// buildDefaultLifecycle builds a build and deploy command lifecycle with reasonable defaults for a non-Java
// project. It uses `gcloud builds submit` for building the samples container image and submitting it to the container
// and `gcloud run deploy` for deploying it to Cloud Run.
func buildDefaultLifecycle(serviceName, gcrURL string) Lifecycle {
	return Lifecycle{
		util.GcloudCommandBuild(
			"builds",
			"submit",
			fmt.Sprintf("--tag=%s", gcrURL),
		),
		util.GcloudCommandBuild(
			"run",
			"deploy",
			serviceName,
			fmt.Sprintf("--image=%s", gcrURL),
			"--platform=managed",
			"--region=us-east4",
		),
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
