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

package gcloud

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"os/exec"
	"strings"
	"unicode"
)

const (
	maxCloudRunServiceNameLen        = 53
	cloudRunServiceNameRandSuffixLen = 10
)

// CloudRunService represents a Cloud Run service and stores its parameters.
type CloudRunService struct {
	Name string
	url  string
}

// Delete calls the external gcloud SDK and deletes the Cloud Run Service associated with the current cloudRunService.
func (s CloudRunService) Delete(sampleDir string) error {
	a := append(util.GcloudCommonFlags, "run", "services", "delete", s.Name, "--platform=managed")
	_, err := util.ExecCommand(exec.Command("gcloud", a...), sampleDir)

	if err != nil {
		return fmt.Errorf("deleting Cloud Run Service: %w", err)
	}

	return nil
}

// URL calls the external gcloud SDK and gets the root URL of the Cloud Run Service associated with the current
// CloudRunService.
func (s *CloudRunService) URL(sampleDir string) (string, error) {
	if s.url != "" {
		return s.url, nil
	}

	a := append(util.GcloudCommonFlags, "run", "--platform=managed", "services", "describe", s.Name,
		"--format=value(status.url)")
	url, err := util.ExecCommand(exec.Command("gcloud", a...), sampleDir)

	if err != nil {
		return "", fmt.Errorf("getting Cloud Run Service URL: %w", err)
	}

	s.url = url
	return url, err
}

// ServiceName generates a Cloud Run service name for the provided sample. It concatenates the sample's name with a
// random alphanumeric string.
func ServiceName(sampleName string) (string, error) {
	randBytes := make([]byte, cloudRunServiceNameRandSuffixLen/2)

	_, err := rand.Read(randBytes)
	if err != nil {
		return "", fmt.Errorf("crypto/rand.Read: %w", err)
	}

	randSuffix := hex.EncodeToString(randBytes)

	l := maxCloudRunServiceNameLen - len(randSuffix) - 1
	sampleName = sampleName[len(sampleName)-l:]
	sampleName = strings.TrimFunc(sampleName, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	return sampleName + "-" + randSuffix, nil
}
