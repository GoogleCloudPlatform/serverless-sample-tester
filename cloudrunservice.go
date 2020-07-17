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
)

// Maximum length of a Cloud Run Service name.
const maxCloudRunServiceNameLen = 53

// cloudRunService represents a Cloud Run service and stores its parameters.
type cloudRunService struct {
	sample *sample
	name   string
	url    string
}

// newCloudRunService returns a new cloudRunService with the provided name.
func newCloudRunService(sample *sample) (s *cloudRunService, err error) {
	n, err := serviceName(sample)
	if err != nil {
		return
	}

	s = &cloudRunService{
		sample: sample,
		name:   n,
	}

	return
}

// delete calls the external gcloud SDK and deletes the Cloud Run Service associated with the current cloudRunService.
func (s *cloudRunService) delete() (err error) {
	_, err = execCommand(gcloudCommandBuild([]string{
		"run",
		"services",
		"delete",
		s.name,
		"--region=us-east4",
		"--platform=managed",
	}))

	return
}

// getURL calls the external gcloud SDK and gets the root URL of the Cloud Run Service associated with the current
// cloudRunService.
func (s *cloudRunService) getURL() (url string, err error) {
	if s.url != "" {
		return s.url, nil
	}

	url, err = execCommand(gcloudCommandBuild([]string{
		"run",
		"--platform=managed",
		"--region=us-east4",
		"services",
		"describe",
		s.name,
		"--format=value(status.url)",
	}))

	return
}

// serviceName generates a Cloud Run service name for the provided sample. It concatenates the sample's name with a
// random 10-character alphanumeric string.
func serviceName(sample *sample) (name string, err error) {
	randBytes := make([]byte, cloudRunServiceNameRandSuffixLen/2)

	_, err = rand.Read(randBytes)
	if err != nil {
		return
	}

	randSuffix := fmt.Sprintf("-%s", hex.EncodeToString(randBytes))
	name = fmt.Sprintf("%s%s", sample.sampleName(len(randSuffix)), randSuffix)

	return
}
