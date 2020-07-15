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
)

// Lifecycle is a representation of a large process that needs to be executed, like building and deploying a sample to
// Cloud Run. Each lifecycle is made up of a list of different phases -- like build, deploy, and post-deploy -- and each
// phase is in turn made up of small goals that act as the steps of each phase.
type Lifecycle struct {
	// id is a unique identifier for this Lifecycle. Not currently used.
	id string

	// phases contains an order list of phases associated with this lifecycle.
	phases []*Phase

	// phaseMap is a map that maps IDs to their associated phase in the Phases slice.
	phaseMap map[string]*Phase
}

// Phase represents one phase of a larger lifecycle. It has a list of goals that make up the actual step-by-step
// execution of the phase in the form of exec.Cmd. In addition, each phase has a string ID that will uniquely identify
// it in the context of its parent Lifecycle.
type Phase struct {
	id    string
	goals []*exec.Cmd
}

// newLifeCycle builds a new Lifecycle given the provided phases and lifecycle id.
func newLifeCycle(id string, phases []*Phase) *Lifecycle {
	lifecycle := Lifecycle{
		id:     id,
		phases: phases,
	}

	lifecycle.phaseMap = make(map[string]*Phase)
	for _, phase := range phases {
		lifecycle.phaseMap[phase.id] = phase
	}

	return &lifecycle
}

// buildBuildDeployLifecycle builds a lifecycle with the following phases: build, deploy, and post-deploy. No goals
// will be attached to any phase.
func buildBuildDeployLifecycle() *Lifecycle {
	return newLifeCycle("build_deploy", []*Phase{
		{
			id:    "build",
			goals: nil,
		},
		{
			id:    "deploy",
			goals: nil,
		},
		{
			id:    "post-deploy",
			goals: nil,
		},
	})
}

// execute executes the goals for each phase.
func (l *Lifecycle) execute() {
	for _, phase := range l.phases {
		log.Printf("Executing %s phase\n", phase.id)

		if phase.goals == nil || len(phase.goals) == 0 {
			continue
		}

		for _, goal := range phase.goals {
			execCommand(goal)
		}
	}
}

// getBuildDeployLifecycle returns a Lifecycle built with reasonable defaults based on whether the sample is java-based
// (has a pom.xml) that doesn't have a Dockerfile or isn't.
func getBuildDeployLifecycle(sample *Sample) *Lifecycle {
	pomPath := fmt.Sprintf("%s/pom.xml", sample.dir)
	dockerfilePath := fmt.Sprintf("%s/Dockerfile", sample.dir)

	_, err := os.Stat(pomPath)
	pomExists := err == nil

	_, err = os.Stat(dockerfilePath)
	dockerfileExists := err == nil

	if pomExists && !dockerfileExists {
		return buildDefaultJavaBuildDeployLifecycle(sample)
	} else {
		return buildDefaultBuildDeployLifecycle(sample)
	}
}

// buildDefaultBuildDeployLifecycle builds a build and deploy command lifecycle with reasonable defaults for a non-Java
// project. It uses `gcloud builds submit` for building the samples container image and submitting it to the container
// and `gcloud run deploy` for deploying it to Cloud Run.
func buildDefaultBuildDeployLifecycle(sample *Sample) *Lifecycle {
	lifecycle := buildBuildDeployLifecycle()

	gcrURL := sample.cloudContainerImage.url()
	lifecycle.phaseMap["build"].goals = []*exec.Cmd{
		gcloudCommandBuild([]string{
			"builds",
			"submit",
			fmt.Sprintf("--tag=%s", gcrURL),
		}),
	}

	lifecycle.phaseMap["deploy"].goals = []*exec.Cmd{
		gcloudCommandBuild([]string{
			"run",
			"deploy",
			sample.cloudRunService.name,
			fmt.Sprintf("--image=%s", gcrURL),
			"--platform=managed",
			"--region=us-east4",
		}),
	}

	return lifecycle
}

// buildDefaultBuildDeployLifecycle builds a build and deploy command lifecycle with reasonable defaults for Java
// samples. It uses `com.google.cloud.tools:jib-maven-plugin:2.0.0:build` for building the samples container image and
// submitting it to the container and `gcloud run deploy` for deploying it to Cloud Run.
func buildDefaultJavaBuildDeployLifecycle(sample *Sample) *Lifecycle {
	lifecycle := buildDefaultBuildDeployLifecycle(sample)

	gcrURL := sample.cloudContainerImage.url()
	lifecycle.phaseMap["build"].goals = []*exec.Cmd{
		exec.Command("mvn",
			"compile",
			"com.google.cloud.tools:jib-maven-plugin:2.0.0:build",
			fmt.Sprintf("-Dimage=%s", gcrURL),
		),
	}

	return lifecycle
}
