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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

// Maximum length of a Cloud Run Service name.
const maxCloudRunServiceNameLen = 53

// CloudRunService represents a Cloud Run Service and holds its name and URL.
type CloudRunService struct {
	// the associated Sample
	sample *Sample

	// the Service's name
	name string

	// the Service's root URL
	url string
}

// newCloudRunService returns a pointer to a new CloudRunService with the provided name.
func newCloudRunService(sample *Sample) *CloudRunService {
	return &CloudRunService{
		sample: sample,
		name:   serviceName(sample),
	}
}

// delete calls the external gcloud SDK and deletes the Cloud Run Service associated with the current CloudRunService.
func (s *CloudRunService) delete() {
	execCommand(gcloudCommandBuild([]string{
		"run",
		"services",
		"delete",
		s.name,
		"--region=us-east4",
		"--platform=managed",
	}))
}

// getURL calls the external gcloud SDK and gets the root URL of the Cloud Run Service associated with the current
// CloudRunService.
func (s *CloudRunService) getURL() string {
	if s.url == "" {
		s.url = execCommand(gcloudCommandBuild([]string{
			"run",
			"--platform=managed",
			"--region=us-east4",
			"services",
			"describe",
			s.name,
			"--format=value(status.url)",
		}))
	}

	return s.url
}

// serviceName generates a Cloud Run service name for the provided sample. It concatenates the sample's name with a
// random 10-character alphanumeric string.
func serviceName(sample *Sample) string {
	randBytes := make([]byte, cloudRunServiceNameRandSuffixLen/2)

	_, err := rand.Read(randBytes)
	if err != nil {
		log.Panicf("Error generating crypto/rand bytes:  %v\n", err)
	}

	randSuffix := fmt.Sprintf("-%s", hex.EncodeToString(randBytes))
	return fmt.Sprintf("%s%s", sample.sampleName(len(randSuffix)), randSuffix)
}
