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
	"os"
	"os/exec"
)

type lifecycle []*exec.Cmd

// execute executes the commands of a lifecycle.
func (l lifecycle) execute() {
	for _, cmd := range l {
		execCommand(cmd)
	}
}

// getLifecycle returns a lifecycle built with reasonable defaults based on whether the sample is java-based
// (has a pom.xml) that doesn't have a Dockerfile or isn't.
func getLifecycle(sample *sample) lifecycle {
	pomPath := fmt.Sprintf("%s/pom.xml", sample.dir)
	dockerfilePath := fmt.Sprintf("%s/Dockerfile", sample.dir)

	_, err := os.Stat(pomPath)
	pomExists := err == nil

	_, err = os.Stat(dockerfilePath)
	dockerfileExists := err == nil

	if pomExists && !dockerfileExists {
		return buildDefaultJavaLifecycle(sample)
	} else {
		return buildDefaultLifecycle(sample)
	}
}

// buildDefaultLifecycle builds a build and deploy command lifecycle with reasonable defaults for a non-Java
// project. It uses `gcloud builds submit` for building the samples container image and submitting it to the container
// and `gcloud run deploy` for deploying it to Cloud Run.
func buildDefaultLifecycle(sample *sample) lifecycle {
	gcrURL := sample.container.url()
	return lifecycle{
		gcloudCommandBuild([]string{
			"builds",
			"submit",
			fmt.Sprintf("--tag=%s", gcrURL),
		}),
		gcloudCommandBuild([]string{
			"run",
			"deploy",
			sample.service.name,
			fmt.Sprintf("--image=%s", gcrURL),
			"--platform=managed",
			"--region=us-east4",
		}),
	}
}

// buildDefaultJavaLifecycle builds a build and deploy command lifecycle with reasonable defaults for Java
// samples. It uses `com.google.cloud.tools:jib-maven-plugin:2.0.0:build` for building the samples container image and
// submitting it to the container and `gcloud run deploy` for deploying it to Cloud Run.
func buildDefaultJavaLifecycle(sample *sample) lifecycle {
	gcrURL := sample.container.url()

	return lifecycle{
		exec.Command("mvn",
			"compile",
			"com.google.cloud.tools:jib-maven-plugin:2.0.0:build",
			fmt.Sprintf("-Dimage=%s", gcrURL),
		),
	}
}
