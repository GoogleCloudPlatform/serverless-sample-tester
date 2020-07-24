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

package sample

import (
	"fmt"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/gcloud"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/lifecycle"
	"github.com/GoogleCloudPlatform/serverless-sample-tester/internal/util"
	"os/exec"
	"strings"
	"unicode"
)

const maxCloudContainerImageTagLen = 53

// Sample represents a Google Cloud Platform sample and associated properties.
type Sample struct {
	Name string

	// The local directory this sample is located in.
	Dir string

	// The cloudRunService this sample will deploy to.
	Service gcloud.CloudRunService

	// The lifecycle for building and deploying this sample to Cloud Run.
	BuildDeployLifecycle lifecycle.Lifecycle

	// The URL location of this sample's build container image in the GCP Container Registry.
	cloudContainerImageURL string
}

// NewSample creates a new sample object for the sample located in the provided local directory.
func NewSample(dir string) (*Sample, error) {
	name := sampleName(dir)

	containerTag, err := cloudContainerImageTag(name)
	if err != nil {
		return nil, fmt.Errorf("[sample.NewSample] generating Container Registry container image tag: %w", err)
	}

	projectID, err := util.ExecCommand(util.GcloudCommandBuild("config", "get-value", "core/project"))
	if err != nil {
		return nil, fmt.Errorf("[sample.NewSample] getting gcloud default project: %w", err)
	}
	cloudContainerImageURL := fmt.Sprintf("gcr.io/%s/%s", projectID, containerTag)

	serviceName, err := gcloud.ServiceName(name)
	if err != nil {
		return nil, fmt.Errorf("[sample.NewSample] generating Cloud Run Service name: %w", err)
	}
	service := gcloud.CloudRunService{Name: serviceName}

	buildDeployLifecycle := lifecycle.GetLifecycle(dir, service.Name, cloudContainerImageURL)

	s := &Sample{
		Name:                   name,
		Dir:                    dir,
		Service:                service,
		BuildDeployLifecycle:   buildDeployLifecycle,
		cloudContainerImageURL: cloudContainerImageURL,
	}
	return s, nil
}

// sampleName computes a sample name for a sample object. Right now, it's defined as a shortened version of the sample's
// local directory. Its length is flexible based on the provided length of a suffix that will be appended to the end of
// the name.
func sampleName(dir string) string {
	n := strings.ReplaceAll(dir, "/", "-")
	return strings.ToLower(n)
}

// DeleteCloudContainerImage deletes the sample's container image off of the Container Registry.
func (s *Sample) DeleteCloudContainerImage() error {
	_, err := util.ExecCommand(util.GcloudCommandBuild(
		"container",
		"images",
		"delete",
		s.cloudContainerImageURL,
	))

	if err != nil {
		return fmt.Errorf("[Sample.DeleteCloudContainerImage] deleting Container Registry container image: %w", err)
	}

	return nil
}

// cloudContainerImageTag creates a container image tag for the provided sample. It concatenates the sample's name
// with a short SHA of the sample repository's HEAD commit.
func cloudContainerImageTag(sampleName string) (string, error) {
	sha, err := util.ExecCommand(exec.Command("git", "rev-parse", "--verify", "--short", "HEAD"))
	if err != nil {
		return "", fmt.Errorf("[sample.cloudContainerImageTag] getting short SHA for sample repository: %w", err)
	}

	l := maxCloudContainerImageTagLen - len(sha) - 1
	sampleName = sampleName[len(sampleName)-l:]
	sampleName = strings.TrimFunc(sampleName, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	tag := sampleName + "-" + sha
	return tag, nil
}
