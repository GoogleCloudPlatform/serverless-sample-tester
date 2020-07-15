package main

import (
	"os"
	"strings"
)

// Sample represents a Google Cloud Platform sample and associated properties.
type Sample struct {
	// The Google Cloud Project ID this sample will deploy to
	googleCloudProject string

	// The local directory this sample is located in
	dir string

	// The CloudRunService this sample will deploy to
	cloudRunService *CloudRunService

	// The CloudContainerImage that represents the location of this sample's build container image in the GCP Container
	// Registry
	cloudContainerImage *CloudContainerImage

	// The Lifecycle for building and deploying this sample to Cloud Run
	buildDeployLifecycle *Lifecycle
}

// newSample creates a new Sample object for the sample located in the provided local directory.
func newSample(dir string) *Sample {
	sample := Sample{
		googleCloudProject: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		dir:                dir,
	}

	sample.cloudRunService = newCloudRunService(&sample)
	sample.cloudContainerImage = newCloudContainerImage(&sample)
	sample.buildDeployLifecycle = getBuildDeployLifecycle(&sample)

	return &sample
}

// sampleName computes a sample name for a sample object. Right now, it's defined as a shortened version of the sample's
// local directory. Its length is flexible based on the provided length of a suffix that will be appended to the end of
// the name.
func (s *Sample) sampleName(suffixLen int) string {
	result := strings.ReplaceAll(s.dir[len(s.dir)-(maxCloudRunServiceNameLen-suffixLen):], "/", "-")

	if result[len(result)-1] == '-' {
		result = result[:len(result)-1]
	}

	if result[0] == '-' {
		result = result[1:]
	}

	return strings.ToLower(result)
}
