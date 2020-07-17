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
func newCloudContainerImage(sample *sample) (c *cloudContainerImage, err error) {
	tag, err := cloudContainerImageTag(sample)
	if err != nil {
		return
	}

	c = &cloudContainerImage{
		sample:       sample,
		containerTag: tag,
	}

	return
}

// newCloudContainerImage creates a new cloudContainerImage for the provided sample.
func (c *cloudContainerImage) url() string {
	return fmt.Sprintf("gcr.io/%s/%s", c.sample.projectID, c.containerTag)
}

// delete deletes the container image off of the Container Registry.
func (c *cloudContainerImage) delete() (err error) {
	_, err = execCommand(gcloudCommandBuild([]string{
		"container",
		"images",
		"delete",
		c.url(),
	}))

	return
}

// cloudContainerImageTag creates a container image tag for the provided sample. It concatenates the sample's name
// with a short SHA of the sample repository's HEAD commit.
func cloudContainerImageTag(sample *sample) (tag string, err error) {
	sha, err := execCommand(exec.Command("git", "rev-parse", "--verify", "--short", "HEAD"))
	if err != nil {
		return
	}

	shortSHASuffix := fmt.Sprintf("-%s", sha)
	tag = fmt.Sprintf("%s%s", sample.sampleName(len(shortSHASuffix)), shortSHASuffix)

	return
}
