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
	"os/exec"
)

// cloudContainerImage represents a sample's container image stored on the GCP Container Registry.
type cloudContainerImage struct {
	// the associated sample
	sample *sample

	// the container image's tag
	containerTag string
}

// newCloudContainerImage creates a new cloudContainerImage for the provided sample.
func newCloudContainerImage(sample *sample) *cloudContainerImage {
	cloudContainerImage := cloudContainerImage{
		sample:       sample,
		containerTag: cloudContainerImageTag(sample),
	}

	return &cloudContainerImage
}

// newCloudContainerImage creates a new cloudContainerImage for the provided sample.
func (c *cloudContainerImage) url() string {
	return fmt.Sprintf("gcr.io/%s/%s", c.sample.projectID, c.containerTag)
}

// delete deletes the container image off of the Container Registry.
func (c *cloudContainerImage) delete() {
	execCommand(gcloudCommandBuild([]string{
		"container",
		"images",
		"delete",
		c.url(),
	}))
}

// cloudContainerImageTag creates a container image tag for the provided sample. It concatenates the sample's name
// with a short SHA of the sample repository's HEAD commit.
func cloudContainerImageTag(sample *sample) string {
	shortSHASuffix := fmt.Sprintf("-%s", execCommand(exec.Command("git", "rev-parse", "--verify", "--short", "HEAD")))
	return fmt.Sprintf("%s%s", sample.sampleName(len(shortSHASuffix)), shortSHASuffix)
}
