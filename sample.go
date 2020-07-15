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
