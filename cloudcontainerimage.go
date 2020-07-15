package main

import (
	"fmt"
	"os/exec"
)

// CloudContainerImage represents a sample's container image stored on the GCP Container Registry.
type CloudContainerImage struct {
	// the associated Sample
	sample *Sample

	// the container image's tag
	containerTag string
}

// newCloudContainerImage creates a new CloudContainerImage for the provided Sample.
func newCloudContainerImage(sample *Sample) *CloudContainerImage {
	cloudContainerImage := CloudContainerImage{
		sample: sample,
		containerTag: cloudContainerImageTag(sample),
	}

	return &cloudContainerImage
}

// newCloudContainerImage creates a new CloudContainerImage for the provided Sample.
func (c *CloudContainerImage) url() string {
	return fmt.Sprintf("gcr.io/%s/%s", c.sample.googleCloudProject, c.containerTag)
}

// delete deletes the container image off of the Container Registry.
func (c *CloudContainerImage) delete() {
	execCommand(gcloudCommandBuild([]string{
		"container",
		"images",
		"delete",
		c.url(),
	}))
}

// cloudContainerImageTag creates a container image tag for the provided sample. It concatenates the sample's name
// with a short SHA of the sample repository's HEAD commit.
func cloudContainerImageTag(sample *Sample) string {
	shortSHASuffix := fmt.Sprintf("-%s", execCommand(exec.Command("git", "rev-parse", "--verify", "--short", "HEAD")))
	return fmt.Sprintf("%s%s", sample.sampleName(len(shortSHASuffix)), shortSHASuffix)
}
